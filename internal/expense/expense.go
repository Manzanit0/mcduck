package expense

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Expense struct {
	ID          uint64
	Date        time.Time
	Amount      float32
	Category    string
	Subcategory string
	UserEmail   string
	ReceiptID   uint64
	Description string
}

// dbExpense is the representation of an expense in the database. For instance,
// the amount is saved as an integer in the DB, but presented to the user as a
// float32.
type dbExpense struct {
	ID          uint64    `db:"id"`
	Date        time.Time `db:"expense_date"`
	Amount      int32     `db:"amount"`
	Category    *string   `db:"category"`
	Subcategory *string   `db:"sub_category"`
	UserEmail   string    `db:"user_email"`
	ReceiptID   *uint64   `db:"receipt_id"`
	Description *string   `db:"description"`
}

func (e Expense) MonthYear() string {
	return NewMonthYear(e.Date)
}

func NewMonthYear(t time.Time) string {
	return t.Format("2006-01")
}

func FindMostRecentTime(expenses []Expense) time.Time {
	var mostRecent time.Time
	for _, e := range expenses {
		if mostRecent.Before(e.Date) {
			mostRecent = e.Date
		}
	}

	return mostRecent
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

type CategoryAggregate struct {
	Category    string
	MonthYear   string
	TotalAmount float32
}

func GetTop3ExpenseCategories(expenses []Expense, monthYear string) []CategoryAggregate {
	var aggregates []CategoryAggregate
	for _, e := range expenses {
		if !strings.EqualFold(e.MonthYear(), monthYear) {
			continue
		}

		if i, aggr, found := findAggregateByCategory(aggregates, e.Subcategory); found {
			// NOTE: there is a lot of converting here. If it ends up being
			// slow; having an intermediate structure which just uses integers
			// and then we do a single final conversion, should help.
			current := ConvertToCents(aggr.TotalAmount)
			total := current + ConvertToCents(e.Amount)
			aggregates[i].TotalAmount = ConvertToDollar(total)
		} else {
			// NOTE: we don't really want to report on empty subcategories since it doesn't provide much value
			if e.Subcategory == "" {
				continue
			}

			aggregates = append(aggregates, CategoryAggregate{
				TotalAmount: e.Amount,
				MonthYear:   monthYear,
				Category:    e.Subcategory,
			})
		}
	}

	sort.Slice(aggregates, func(i, j int) bool {
		return aggregates[i].TotalAmount > aggregates[j].TotalAmount
	})

	if len(aggregates) > 3 {
		return aggregates[:3]
	}

	return aggregates
}

func findAggregateByCategory(aggregates []CategoryAggregate, category string) (int, CategoryAggregate, bool) {
	for i, a := range aggregates {
		if strings.EqualFold(a.Category, category) {
			return i, a, true
		}
	}

	return 0, CategoryAggregate{}, false
}

func SortByDate(expenses []Expense) {
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Date.After(expenses[j].Date)
	})
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

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type ExpensesBatch struct {
	Records   []Expense
	UserEmail string
}

func (r *Repository) CreateExpenses(ctx context.Context, e ExpensesBatch) error {
	return CreateExpenses(ctx, r.db, e)
}

type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func CreateExpenses(ctx context.Context, tx QueryExecutor, e ExpensesBatch) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Insert("expenses").Columns(
		"user_email",
		"expense_date",
		"amount",
		"category",
		"sub_category",
		"description",
		"receipt_id",
	)

	for _, expense := range e.Records {
		builder = builder.Values(
			e.UserEmail,
			expense.Date,
			ConvertToCents(expense.Amount),
			expense.Category,
			expense.Subcategory,
			expense.Description,
			expense.ReceiptID,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	return nil
}

func (r *Repository) FindExpense(ctx context.Context, id int64) (*Expense, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.
		Select("id, expense_Date, amount, category, sub_category, user_email", "receipt_id").
		From("expenses").
		Where(sq.Eq{"id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to build query: %w", err)
	}

	var out dbExpense
	err = r.db.GetContext(ctx, &out, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	expense := toDomainExpense(out)
	return &expense, nil
}

type UpdateExpenseRequest struct {
	ID          int64
	Date        *time.Time
	Amount      *float32
	Category    *string
	Subcategory *string
	Description *string
	ReceiptID   *uint64
}

func (r *Repository) UpdateExpense(ctx context.Context, e UpdateExpenseRequest) error {
	var shouldUpdate bool

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update("expenses").Where(sq.Eq{"id": e.ID})

	if e.Amount != nil {
		builder = builder.Set("amount", ConvertToCents(*e.Amount))
		shouldUpdate = true
	}

	if e.Category != nil {
		builder = builder.Set("category", *e.Category)
		shouldUpdate = true
	}

	if e.Subcategory != nil {
		builder = builder.Set("sub_category", *e.Subcategory)
		shouldUpdate = true
	}

	if e.Description != nil {
		builder = builder.Set("description", *e.Description)
		shouldUpdate = true
	}

	if e.ReceiptID != nil {
		builder = builder.Set("receipt_id", *e.ReceiptID)
		shouldUpdate = true
	}

	if e.Date != nil {
		builder = builder.Set("expense_date", *e.Date)
		shouldUpdate = true
	}

	if !shouldUpdate {
		return nil
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	return nil
}

type CreateExpenseRequest struct {
	UserEmail string
	Date      time.Time
	Amount    float32
	ReceiptID *uint64
}

func (r *Repository) CreateExpense(ctx context.Context, e CreateExpenseRequest) (int64, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.
		Insert("expenses").
		Columns("user_email", "amount, expense_date", "receipt_id").
		Values(e.UserEmail, ConvertToCents(e.Amount), e.Date, e.ReceiptID).
		Suffix("RETURNING \"id\"")

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("unable to build query: %w", err)
	}

	record := struct {
		ID int64 `db:"id"`
	}{}

	err = r.db.GetContext(ctx, &record, query, args...)
	if err != nil {
		return 0, fmt.Errorf("unable to execute query: %w", err)
	}

	return record.ID, nil
}

func (r *Repository) DeleteExpense(ctx context.Context, id int64) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Delete("expenses").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	return nil
}

func (r *Repository) ListExpenses(ctx context.Context, email string) ([]Expense, error) {
	var expenses []dbExpense
	err := r.db.SelectContext(ctx, &expenses, `SELECT id, amount, expense_date, category, sub_category, description, receipt_id FROM expenses WHERE user_email = $1 ORDER BY expense_date desc`, email)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	var expensesList []Expense
	for _, expense := range expenses {
		expensesList = append(expensesList, toDomainExpense(expense))
	}

	return expensesList, nil
}

func (r *Repository) ListExpensesForReceipt(ctx context.Context, receiptID uint64) ([]Expense, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "expense_date", "amount", "category", "sub_category", "description").
		From("expenses").
		Where(sq.Eq{"receipt_id": receiptID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var expenses []dbExpense
	err = r.db.SelectContext(ctx, &expenses, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	var expensesList []Expense
	for _, expense := range expenses {
		expensesList = append(expensesList, toDomainExpense(expense))
	}

	return expensesList, nil
}

func ConvertToCents(amount float32) int32 {
	return int32(math.Round(float64(amount * 100)))
}

func ConvertToDollar(cents int32) float32 {
	if cents == 0 {
		return float32(0)
	}

	return float32(cents) / 100
}

func toDomainExpense(expense dbExpense) Expense {
	e := Expense{
		ID:        expense.ID,
		Date:      expense.Date,
		Amount:    ConvertToDollar(expense.Amount),
		UserEmail: expense.UserEmail,
	}

	if expense.Category != nil {
		e.Category = *expense.Category
	}

	if expense.Subcategory != nil {
		e.Subcategory = *expense.Subcategory
	}

	if expense.ReceiptID != nil {
		e.ReceiptID = *expense.ReceiptID
	}

	if expense.Description != nil {
		e.Description = *expense.Description
	}

	return e
}
