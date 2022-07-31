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
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/assets", "templates/assets")

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

	if err := r.Run(); err != nil {
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
	nowMonthYear := time.Now().Format("2006-01")
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
