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

	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/pkg/auth"
)

var chartColours = []string{
	"rgba(255, 99, 132)",
	"rgba(255, 159, 64)",
	"rgba(255, 205, 86)",
	"rgba(75, 192, 192)",
	"rgba(54, 162, 235)",
	"rgba(153, 102, 255)",
	"rgba(201, 203, 207)",
}

var chartBackgroundColours = []string{
	"rgba(255, 99, 132, 0.2)",
	"rgba(255, 159, 64, 0.2)",
	"rgba(255, 205, 86, 0.2)",
	"rgba(75, 192, 192, 0.2)",
	"rgba(54, 162, 235, 0.2)",
	"rgba(153, 102, 255, 0.2)",
	"rgba(201, 203, 207, 0.2)",
}

type ChartData struct {
	Title    string
	Labels   []string
	Datasets []Dataset
}

type Dataset struct {
	Label            string
	BorderColour     string
	BackgroundColour string
	Hidden           bool
	Data             []string
}

type DashboardController struct {
	Expenses   *expense.Repository
	SampleData []expense.Expense
}

func (d *DashboardController) LiveDemo(c *gin.Context) {
	expenses := d.SampleData

	expense.SortByDate(expenses)

	mostRecent := expense.FindMostRecentTime(expenses)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	categoryTotals := expense.CalculateTotalsPerCategory(expenses)
	categoryLabels := getSecondClassifier(categoryTotals)
	categoryChartData := buildChartData(categoryLabels, categoryTotals)

	subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
	subcategoryLabels := getSecondClassifier(subcategoryTotals)
	subcategoryChartData := buildChartData(subcategoryLabels, subcategoryTotals)

	// Since this is for public demoing, we might as well show-off the whole data
	// off the bat.
	for i := range categoryChartData.Datasets {
		categoryChartData.Datasets[i].Hidden = false
	}

	for i := range subcategoryChartData.Datasets {
		subcategoryChartData.Datasets[i].Hidden = false
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":        mostRecent.Format("January 2006"),
		"NoExpenses":             len(expenses) == 0,
		"Categories":             categoryLabels,
		"CategoriesChartData":    categoryChartData,
		"SubCategories":          subcategoryLabels,
		"SubCategoriesChartData": subcategoryChartData,
		"TopCategories":          expense.GetTop3ExpenseCategories(expenses, mostRecentMonthYear),
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

	expense.SortByDate(expenses)

	mostRecent := expense.FindMostRecentTime(expenses)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	categoryTotals := expense.CalculateTotalsPerCategory(expenses)
	categoryLabels := getSecondClassifier(categoryTotals)
	categoryChartData := buildChartData(categoryLabels, categoryTotals)

	var subcategoryCharts []ChartData
	for cat, subcats := range GroupSubcategoriesByCategory(expenses) {
		filtered := FilterByCategory(expenses, cat)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(filtered)
		subcategoryChartData := buildChartData(subcats, subcategoryTotals)

		subcategoryChartData.Title = cat

		subcategoryCharts = append(subcategoryCharts, subcategoryChartData)
	}

	totalSpendsArr := TotalSpendLastThreeMonths(expenses)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":        mostRecent.Format("January 2006"),
		"NoExpenses":             len(expenses) == 0,
		"Categories":             categoryLabels,
		"CategoriesChartData":    categoryChartData,
		"SubcategoriesChartData": subcategoryCharts,
		"TopCategories":          expense.GetTop3ExpenseCategories(expenses, mostRecentMonthYear),
		"TotalSpends":            totalSpendsArr,
		"User":                   user,
	})
}

type MonthlySpend struct {
	date      time.Time
	amount    float32
	MonthYear string
	Amount    string
}

func TotalSpendLastThreeMonths(expenses []expense.Expense) []*MonthlySpend {
	totalSpends := map[string]*MonthlySpend{}
	for i := range expenses {
		if isOlderThanLastThreeMonths(expenses[i].Date) {
			continue
		}

		key := expenses[i].Date.Format("January 2006")
		val, ok := totalSpends[key]
		if !ok {
			totalSpends[key] = &MonthlySpend{
				date:      expenses[i].Date,
				MonthYear: key,
				amount:    expenses[i].Amount,
				Amount:    fmt.Sprintf("%.2f", expenses[i].Amount),
			}
		} else {
			val.amount += expenses[i].Amount
			val.Amount = fmt.Sprintf("%.2f", val.amount)
		}
	}

	sortedTotalSpends := []*MonthlySpend{}
	for _, a := range totalSpends {
		sortedTotalSpends = append(sortedTotalSpends, a)
	}

	sort.Slice(sortedTotalSpends, func(i, j int) bool {
		return sortedTotalSpends[i].date.Before(sortedTotalSpends[j].date)
	})

	return sortedTotalSpends
}

func isOlderThanLastThreeMonths(d time.Time) bool {
	// 15th of March 2022-> 15th of December 2022
	year, month, _ := time.Now().AddDate(0, -2, 0).Date()

	// 1st of December 2022
	beginningOf3MonthsAgo := time.Date(year, month, 1, 0, 0, 0, 0, time.Now().Location())

	return d.Before(beginningOf3MonthsAgo)
}

func FilterByCategory(list []expense.Expense, cat string) []expense.Expense {
	var filtered []expense.Expense
	for i := range list {
		if list[i].Category == cat {
			filtered = append(filtered, list[i])
		}
	}

	return filtered
}

func GroupSubcategoriesByCategory(list []expense.Expense) map[string][]string {
	m := map[string]map[string]bool{}
	for _, e := range list {
		if _, ok := m[e.Category]; !ok {
			m[e.Category] = map[string]bool{}
		}

		m[e.Category][e.Subcategory] = true
	}

	mm := map[string][]string{}
	for k, v := range m {
		if _, ok := mm[k]; !ok {
			mm[k] = []string{}
		}

		for s := range v {
			mm[k] = append(mm[k], s)
		}
	}

	return mm
}

func buildChartData(labels []string, totals map[string]map[string]float32) ChartData {
	var datasets []Dataset
	for monthYear, amountsByCategory := range totals { // totalsByMonth[monthYear][expense.Category] += expense.Amount
		var data []string
		for _, label := range labels {
			if amount, ok := amountsByCategory[label]; ok {
				data = append(data, fmt.Sprintf("%.2f", amount))
			} else {
				data = append(data, "0.00")
			}
		}

		datasets = append(datasets, Dataset{
			Label:  monthYear,
			Data:   data,
			Hidden: true,
		})
	}

	// FIXME: very naive sort. We would want to do a time comparison.
	sort.Slice(datasets, func(i, j int) bool {
		return datasets[i].Label < datasets[j].Label
	})

	// By default we only show the current month.
	datasets[len(datasets)-1].Hidden = false

	for i := range datasets {
		datasets[i].BorderColour = chartColours[i]
		datasets[i].BackgroundColour = chartBackgroundColours[i]
	}

	return ChartData{
		Labels:   labels,
		Datasets: datasets,
	}
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

	expense.SortByDate(expenses)

	mostRecent := expense.FindMostRecentTime(expenses)
	mostRecentMonthYear := expense.NewMonthYear(mostRecent)

	categoryTotals := expense.CalculateTotalsPerCategory(expenses)
	categoryLabels := getSecondClassifier(categoryTotals)
	categoryChartData := buildChartData(categoryLabels, categoryTotals)

	subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
	subcategoryLabels := getSecondClassifier(subcategoryTotals)
	subcategoryChartData := buildChartData(subcategoryLabels, subcategoryTotals)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"PrettyMonthYear":        mostRecent.Format("January 2006"),
		"NoExpenses":             len(expenses) == 0,
		"Categories":             categoryLabels,
		"CategoriesChartData":    categoryChartData,
		"SubCategories":          subcategoryLabels,
		"SubCategoriesChartData": subcategoryChartData,
		"TopCategories":          expense.GetTop3ExpenseCategories(expenses, mostRecentMonthYear),
		"User":                   user,
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

func readExpensesFromCSV(filename string) ([]expense.Expense, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return expense.FromCSV(f)
}
