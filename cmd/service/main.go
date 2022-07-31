package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/expense"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{})
	})

	r.GET("/", func(c *gin.Context) {
		expenses, err := readExpensesFromCSV("../../example_input.csv")
		if err != nil {
			log.Fatal(err)
		}

		categoryTotals := expense.CalculateTotalsPerCategory(expenses)
		subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
		mom := expense.CalculateMonthOverMonthTotals(expenses)
		labels, amountsByCategory := getMOMData(mom)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Categories":         getAllTimeCategories(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getAllTimeCategories(subcategoryTotals),
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
			"Categories":         getAllTimeCategories(categoryTotals),
			"CategoryAmounts":    getCurrentMonthAmounts(categoryTotals),
			"SubCategories":      getAllTimeCategories(subcategoryTotals),
			"SubCategoryAmounts": getCurrentMonthAmounts(subcategoryTotals),
			"MOMLabels":          labels,
			"MOMData":            amountsByCategory,
		})
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

func getAllTimeCategories(calculations map[string]map[string]float32) []string {
	// get all unique categoriesMap
	categoriesMap := make(map[string]bool)
	for _, amountByCategory := range calculations {
		for category := range amountByCategory {
			if ok := categoriesMap[category]; !ok {
				categoriesMap[category] = true
			}
		}
	}

	// first get the categories header
	categoriesSlice := make([]string, len(categoriesMap))
	i := 0
	for category := range categoriesMap {
		categoriesSlice[i] = category
		i++
	}

	return categoriesSlice
}

func getCurrentMonthAmounts(calculations map[string]map[string]float32) []string {
	categories := getAllTimeCategories(calculations)
	amounts := []string{}
	nowMonthYear := time.Now().Format("2006-01")
	for _, category := range categories {
		if v := calculations[nowMonthYear][category]; v != 0 {
			amounts = append(amounts, fmt.Sprintf("%.2f", v))
		} else {
			amounts = append(amounts, "0.00")
		}
	}

	return amounts
}

func getMOMData(calculations map[string]map[string]float32) ([]string, map[string][]string) {
	monthsMap := map[string]bool{}
	amountsPerMonthByCategory := map[string][]string{}

	i := 0
	for category := range calculations {
		for month, amount := range calculations[category] {
			monthsMap[month] = true
			amountsPerMonthByCategory[category] = append(amountsPerMonthByCategory[category], fmt.Sprintf("%.2f", amount))
		}
		i++
	}

	var monthsSlice []string
	for month := range monthsMap {
		monthsSlice = append(monthsSlice, month)
	}

	sort.Strings(monthsSlice)

	return monthsSlice, amountsPerMonthByCategory
}

func readExpensesFromCSV(filename string) ([]expense.Expense, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = 4
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	expenses, err := expense.NewExpenses(data[1:])
	if err != nil {
		return nil, err
	}

	return expenses, nil
}
