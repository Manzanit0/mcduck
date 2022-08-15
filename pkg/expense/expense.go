package expense

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Expense struct {
	ID          int64     `db:"id"`
	Date        time.Time `db:"expense_date"`
	Amount      float32   `db:"amount"`
	Category    string    `db:"category"`
	Subcategory string    `db:"sub_category"`
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

type ExpensesBatch struct {
	Records   []Expense
	UserEmail string
}

func CreateExpenses(ctx context.Context, db *sqlx.DB, e ExpensesBatch) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Insert("expenses").Columns("user_email", "expense_date", "amount", "category", "sub_category")
	for _, expense := range e.Records {
		builder = builder.Values(e.UserEmail, expense.Date, ConvertToCents(expense.Amount), expense.Category, expense.Subcategory)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	return nil
}

func ListExpenses(ctx context.Context, db *sqlx.DB, email string) ([]Expense, error) {
	var expenses []Expense
	err := db.SelectContext(ctx, &expenses, `SELECT id, amount, expense_date, category, sub_category FROM expenses WHERE user_email = $1`, email)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	return expenses, nil
}

func ConvertToCents(amount float32) int32 {
	return int32(math.Round(float64(amount * 100)))
}
