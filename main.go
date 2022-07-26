package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/manzanit0/mcduck/pkg/term"
)

func main() {
	// open file
	f, err := os.Open("./example_input.csv")
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = 4
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// skip header
	expenses, err := mapToExpenses(data[1:])
	if err != nil {
		log.Fatal(err)
	}

	categoryTotals := calculateTotalsPerCategory(expenses)
	term.Render("category", categoryTotals)

	subcategoryTotals := calculateTotalsPerSubCategory(expenses)
	term.Render("sub-category", subcategoryTotals)
}

func calculateTotalsPerCategory(expenses []Expense) map[string]map[string]float32 {
	totalsByMonth := make(map[string]map[string]float32)
	for _, expense := range expenses {
		monthYear := expense.Date.Format("2006-01")
		if _, ok := totalsByMonth[monthYear]; !ok {
			totalsByMonth[monthYear] = make(map[string]float32)
		}

		totalsByMonth[monthYear][expense.Category] += expense.Amount
	}

	return totalsByMonth
}

func calculateTotalsPerSubCategory(expenses []Expense) map[string]map[string]float32 {
	totalsByMonth := make(map[string]map[string]float32)
	for _, expense := range expenses {
		monthYear := expense.Date.Format("2006-01")
		if _, ok := totalsByMonth[monthYear]; !ok {
			totalsByMonth[monthYear] = make(map[string]float32)
		}

		totalsByMonth[monthYear][expense.Subcategory] += expense.Amount
	}

	return totalsByMonth
}

func mapToExpenses(data [][]string) ([]Expense, error) {
	expenses := make([]Expense, len(data))
	for k, rows := range data {
		date, err := time.Parse("2006-01-02", rows[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %s for row %d: %w", rows[0], k, err)
		}

		amount, err := strconv.ParseFloat(rows[1], 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount %s for row %d: %w", rows[1], k, err)
		}

		expenses[k] = Expense{
			Date:        date,
			Amount:      float32(amount),
			Category:    rows[2],
			Subcategory: rows[3],
		}
	}

	return expenses, nil
}

type Expense struct {
	Date        time.Time
	Amount      float32
	Category    string
	Subcategory string
}

func (e Expense) MonthYear() string {
	return e.Date.Format("2006-01")
}
