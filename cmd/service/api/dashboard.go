package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/pkg/auth"
)

type DashboardController struct {
	Expenses   *expense.Repository
	SampleData []expense.Expense
}

func (d *DashboardController) LiveDemo(c *gin.Context) {
	categoryTotals := expense.CalculateTotalsPerCategory(d.SampleData)
	subcategoryTotals := expense.CalculateTotalsPerSubCategory(d.SampleData)
	mom := expense.CalculateMonthOverMonthTotals(d.SampleData)
	labels, amountsByCategory := getMOMData(mom)

	mostRecent := expense.FindMostRecentTime(d.SampleData)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":    mostRecent.Format("January 2006"),
		"Categories":         getSecondClassifier(categoryTotals),
		"CategoryAmounts":    getAmountsForMonth(mostRecentMonthYear, categoryTotals),
		"SubCategories":      getSecondClassifier(subcategoryTotals),
		"SubCategoryAmounts": getAmountsForMonth(mostRecentMonthYear, subcategoryTotals),
		"TopCategories":      expense.GetTop3ExpenseCategories(d.SampleData, mostRecentMonthYear),
		"MOMLabels":          labels,
		"MOMData":            amountsByCategory,
	})
}

func (d *DashboardController) Dashboard(c *gin.Context) {
	user := auth.GetUserEmail(c)

	expenses, err := d.Expenses.ListExpenses(c.Request.Context(), user)
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

	mostRecent := expense.FindMostRecentTime(expenses)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	topCategories := expense.GetTop3ExpenseCategories(expenses, mostRecentMonthYear)

	// FIXME: if the subcategory is empty, then it displays an empty card.

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":    mostRecent.Format("January 2006"),
		"NoExpenses":         len(expenses) == 0,
		"Categories":         getSecondClassifier(categoryTotals),
		"CategoryAmounts":    getAmountsForMonth(mostRecentMonthYear, categoryTotals),
		"SubCategories":      getSecondClassifier(subcategoryTotals),
		"SubCategoryAmounts": getAmountsForMonth(mostRecentMonthYear, subcategoryTotals),
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
		err = d.Expenses.CreateExpenses(c.Request.Context(), expense.ExpensesBatch{UserEmail: user, Records: expenses})
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}
	}

	categoryTotals := expense.CalculateTotalsPerCategory(expenses)
	subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
	mom := expense.CalculateMonthOverMonthTotals(expenses)
	labels, amountsByCategory := getMOMData(mom)

	mostRecent := expense.FindMostRecentTime(expenses)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":    mostRecent.Format("January 2006"),
		"Categories":         getSecondClassifier(categoryTotals),
		"CategoryAmounts":    getAmountsForMonth(mostRecentMonthYear, categoryTotals),
		"SubCategories":      getSecondClassifier(subcategoryTotals),
		"SubCategoryAmounts": getAmountsForMonth(mostRecentMonthYear, subcategoryTotals),
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

func getAmountsForMonth(monthYear string, calculations map[string]map[string]float32) []string {
	uniqueCategories := getSecondClassifier(calculations)
	amounts := []string{}
	for _, category := range uniqueCategories {
		amounts = append(amounts, fmt.Sprintf("%.2f", calculations[monthYear][category]))
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
