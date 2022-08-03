package expense

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Expense struct {
	Date        time.Time
	Amount      float32
	Category    string
	Subcategory string
}

func (e Expense) MonthYear() string {
	return e.Date.Format("2006-01")
}

func CalculateTotalsPerCategory(expenses []Expense) map[string]map[string]float32 {
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

func CalculateTotalsPerSubCategory(expenses []Expense) map[string]map[string]float32 {
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

func CalculateMonthOverMonthTotals(expenses []Expense) map[string]map[string]float32 {
	totalsByCategory := make(map[string]map[string]float32)
	for _, expense := range expenses {
		if _, ok := totalsByCategory[expense.Category]; !ok {
			totalsByCategory[expense.Category] = make(map[string]float32)
		}

		monthYear := expense.Date.Format("2006-01")
		totalsByCategory[expense.Category][monthYear] += expense.Amount
	}

	return totalsByCategory
}

func NewExpenses(data [][]string) ([]Expense, error) {
	expenses := make([]Expense, len(data))
	for k, rows := range data {
		date, err := time.Parse("2006-01-02", rows[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %s for row %d: %w", rows[0], k, err)
		}

		// Americans use a dot as a decimal operator, but Spain uses a comma.
		// Support both anyways.
		if strings.ContainsRune(rows[1], ',') {
			rows[1] = strings.ReplaceAll(rows[1], ",", ".")
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

func FromCSV(r io.Reader) ([]Expense, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	csvReader := csv.NewReader(tee)
	csvReader.TrimLeadingSpace = true
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = 4

	data, err := csvReader.ReadAll()
	if err != nil {
		csvReader = csv.NewReader(&buf)
		csvReader.Comma = ','
		csvReader.TrimLeadingSpace = true
		csvReader.FieldsPerRecord = 4

		data, err = csvReader.ReadAll()
		if err != nil {
			return nil, err
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	expenses, err := NewExpenses(data[1:])
	if err != nil {
		return nil, err
	}

	return expenses, nil
}
