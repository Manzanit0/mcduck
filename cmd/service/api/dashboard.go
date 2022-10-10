package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/expense"
)

type DashboardController struct {
	DB         *sqlx.DB
	SampleData []expense.Expense
}

func (d *DashboardController) LiveDemo(c *gin.Context) {
	categoryTotals := expense.CalculateTotalsPerCategory(d.SampleData)
	subcategoryTotals := expense.CalculateTotalsPerSubCategory(d.SampleData)
	mom := expense.CalculateMonthOverMonthTotals(d.SampleData)
	labels, amountsByCategory := getMOMData(mom)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Categories":         getSecondClassifier(categoryTotals),
		"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
		"SubCategories":      getSecondClassifier(subcategoryTotals),
		"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
		"MOMLabels":          labels,
		"MOMData":            amountsByCategory,
	})
}

func (d *DashboardController) Dashboard(c *gin.Context) {
	user := auth.GetUserEmail(c)

	expenses, err := expense.ListExpenses(c.Request.Context(), d.DB, user)
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

	expense.SortByDate(expenses)
	topCategories := expense.GetTop3ExpenseCategories(expenses, expenses[0].MonthYear())
	// FIXME: if the subcategory is empty, then it displays an empty card.

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"NoExpenses":         len(expenses) == 0,
		"Categories":         getSecondClassifier(categoryTotals),
		"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
		"SubCategories":      getSecondClassifier(subcategoryTotals),
		"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
		"TopCategories":      topCategories,
		"MOMLabels":          labels,
		"MOMData":            amountsByCategory,
		"User":               user,
	})
}

func (d *DashboardController) UploadExpenses(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "get form error: %s", err.Error())
		return
	}

	if len(form.File["files"]) == 0 {
		c.String(http.StatusBadRequest, "no files uploaded")
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
	user := auth.GetUserEmail(c)
	if user != "" {
		err = expense.CreateExpenses(c.Request.Context(), d.DB, expense.ExpensesBatch{UserEmail: user, Records: expenses})
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
