package main

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/manzanit0/mcduck/pkg/expense"
)

//go:embed templates/*.html
var templates embed.FS

//go:embed assets/*.ico assets/*.css
var assets embed.FS

//go:embed sample_data.csv
var sampleData embed.FS

func main() {
	t, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	db, err := sql.Open("pgx", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGDATABASE")))
	if err != nil {
		log.Fatalf("unable to open db conn: %s", err.Error())
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("error closing db connection: %s\n", err.Error())
		}
	}()

	r.Use(func(c *gin.Context) {
		c.Set("db", sqlx.NewDb(db, "postgres"))
		c.Next()
	})

	r.Use(CookieAuthMiddleware)
	r.GET("/register", GetRegisterForm)
	r.POST("/register", RegisterUser)
	r.GET("/login", GetLoginForm)
	r.POST("/login", LoginUser)
	r.GET("/signout", Signout)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"User": GetUserEmail(c),
		})
	})

	r.GET("/live_demo", func(c *gin.Context) {
		expenses, err := readSampleData()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Categories":         getSecondClassifier(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getSecondClassifier(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
		})
	})

	r.GET("/dashboard", func(c *gin.Context) {
		user := GetUserEmail(c)
		if user == "" {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "401: Unauthorized"})
			return
		}

		db, ok := c.Get("db")
		if !ok {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "500: Internal Server Error"})
			return
		}

		dbx := db.(*sqlx.DB)
		expenses, err := expense.ListExpenses(c.Request.Context(), dbx, user)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		if len(expenses) == 0 {
			expenses = []expense.Expense{}
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Categories":         getSecondClassifier(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getSecondClassifier(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
			"User":               user,
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form error: %s", err.Error())
			return
		}

		file := form.File["files"][0]
		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusInternalServerError, "upload file error: %s", err.Error())
			return
		}

		expenses, err := readExpensesFromCSV(filename)
		if err != nil {
			c.String(http.StatusInternalServerError, "file parsing error: %s", err.Error())
			return
		}

		// If the user is logged in, save those upload expenses
		user := GetUserEmail(c)
		if user != "" {
			db, ok := c.Get("db")
			if !ok {
				c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
				return
			}

			dbx := db.(*sqlx.DB)
			err = expense.CreateExpenses(c.Request.Context(), dbx, expense.ExpensesBatch{UserEmail: user, Records: expenses})
			if err != nil {
				c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
				return
			}
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Categories":         getSecondClassifier(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getSecondClassifier(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
			"User":               user,
		})
	})

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}

func getSecondClassifier(calculations map[string]map[string]float32) []string {
	classifierMap := map[string]bool{}
	classifierSlice := []string{}
	for _, amountByClassifier := range calculations {
		for secondClassifier := range amountByClassifier {
			if ok := classifierMap[secondClassifier]; !ok {
				classifierMap[secondClassifier] = true
				classifierSlice = append(classifierSlice, secondClassifier)
			}
		}
	}

	sort.Strings(classifierSlice)
	return classifierSlice
}

func getCurrentMonthAmounts(calculations map[string]map[string]float32) []string {
	uniqueCategories := getSecondClassifier(calculations)
	amounts := []string{}
	nowMonthYear := time.Now().AddDate(0, -1, 0).Format("2006-01")
	for _, category := range uniqueCategories {
		amounts = append(amounts, fmt.Sprintf("%.2f", calculations[nowMonthYear][category]))
	}

	return amounts
}

func getMOMData(calculations map[string]map[string]float32) ([]string, map[string][]string) {
	uniqueMonths := getSecondClassifier(calculations)

	amountsPerMonthByCategory := map[string][]string{}

	for category := range calculations {
		for _, month := range uniqueMonths {
			amount := calculations[category][month]
			amountsPerMonthByCategory[category] = append(amountsPerMonthByCategory[category], fmt.Sprintf("%.2f", amount))
		}
	}

	return uniqueMonths, amountsPerMonthByCategory
}

func readExpensesFromCSV(filename string) ([]expense.Expense, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return expense.FromCSV(f)
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
