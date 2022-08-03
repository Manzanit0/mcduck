package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

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

	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{})
	})

	r.GET("/", func(c *gin.Context) {
		expenses, err := readSampleData()
		if err != nil {
			log.Fatal(err)
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Categories":         getSecondClassifier(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getSecondClassifier(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}

		file := form.File["files"][0]
		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			return
		}

		expenses, err := readExpensesFromCSV(filename)
		if err != nil {
			log.Fatal(err)
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Categories":         getSecondClassifier(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getSecondClassifier(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
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
