package receipt

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/manzanit0/isqlx"
	"github.com/manzanit0/mcduck/internal/expense"
)

type Receipt struct {
	ID            int64
	PendingReview bool
	Image         []byte
	Vendor        string
	UserEmail     string
	Date          time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type dbReceipt struct {
	ID            int64   `db:"id"`
	PendingReview bool    `db:"pending_review"`
	Image         []byte  `db:"receipt_image"`
	UserEmail     string  `db:"user_email"`
	Vendor        *string `db:"vendor"`

	Date      time.Time `db:"receipt_date"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *dbReceipt) MapReceipt() *Receipt {
	var vendor string
	if r.Vendor != nil {
		vendor = *r.Vendor
	}
	return &Receipt{
		ID:            r.ID,
		PendingReview: r.PendingReview,
		Image:         r.Image,
		Date:          r.Date,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Vendor:        vendor,
		UserEmail:     r.UserEmail,
	}
}

type Repository struct {
	dbx isqlx.DBX
}

func NewRepository(dbx isqlx.DBX) *Repository {
	return &Repository{dbx: dbx}
}

type CreateReceiptRequest struct {
	Amount      float64
	Description string
	Vendor      string
	Image       []byte
	Date        time.Time
	Email       string
}

func (r *Repository) CreateReceipt(ctx context.Context, input CreateReceiptRequest) (*Receipt, error) {
	if len(input.Image) == 0 {
		return nil, fmt.Errorf("empty receipt")
	}

	txn, err := r.dbx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer txn.TxClose(ctx)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.
		Insert("receipts").
		Columns("receipt_image", "pending_review", "user_email", "receipt_date", "vendor").
		Values(input.Image, true, input.Email, input.Date, input.Vendor).
		Suffix("RETURNING \"id\"")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to build query: %w", err)
	}

	var record dbReceipt
	err = txn.GetContext(ctx, &record, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	if input.Amount > 0 {
		e := expense.ExpensesBatch{
			UserEmail: input.Email,
			Records: []expense.Expense{{
				ReceiptID:   uint64(record.ID),
				Date:        input.Date,
				Amount:      float32(input.Amount),
				UserEmail:   input.Email,
				Description: input.Description,
				Category:    "Receipt Upload",
			}},
		}

		err = expense.CreateExpenses(ctx, txn, e)
		if err != nil {
			return nil, fmt.Errorf("unable to insert expenses: %w", err)
		}
	}

	err = txn.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return record.MapReceipt(), nil
}

type UpdateReceiptRequest struct {
	ID            int64
	Vendor        *string
	PendingReview *bool
	Date          *time.Time
}

func (r *Repository) UpdateReceipt(ctx context.Context, e UpdateReceiptRequest) error {
	var shouldUpdate bool
	var shouldUpdateExpenseDates bool

	txn, err := r.dbx.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer txn.TxClose(ctx)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update("receipts").Where(sq.Eq{"id": e.ID})

	if e.Vendor != nil {
		builder = builder.Set("vendor", *e.Vendor)
		shouldUpdate = true
	}

	if e.PendingReview != nil {
		builder = builder.Set("pending_review", *e.PendingReview)
		shouldUpdate = true
	}

	if e.Date != nil {
		builder = builder.Set("receipt_date", *e.Date)
		shouldUpdate = true
		shouldUpdateExpenseDates = true
	}

	if !shouldUpdate {
		return nil
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("compile receipts query: %w", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	if shouldUpdateExpenseDates {
		query, args, err = psql.Update("expenses").Where(sq.Eq{"receipt_id": e.ID}).Set("expense_date", *e.Date).ToSql()
		if err != nil {
			return fmt.Errorf("compile expenses query: %w", err)
		}

		_, err = txn.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("execute expenses query: %w", err)
		}
	}

	err = txn.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (r *Repository) ListReceipts(ctx context.Context, email string) ([]Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "receipt_date").
		From("receipts").
		Where(sq.Eq{"user_email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipts []dbReceipt
	err = r.dbx.SelectContext(ctx, &receipts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	var domainReceipts []Receipt
	for _, receipt := range receipts {
		domainReceipts = append(domainReceipts, *receipt.MapReceipt())
	}

	return domainReceipts, nil
}

func (r *Repository) ListReceiptsCurrentMonth(ctx context.Context, email string) ([]Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "receipt_date").
		From("receipts").
		Where(sq.And{
			sq.Eq{"user_email": email},
			sq.Expr("receipt_date >= date_trunc('month',current_date)"),
			sq.Expr("receipt_date < date_trunc('month',current_date) + INTERVAL '1' MONTH"),
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipts []dbReceipt
	err = r.dbx.SelectContext(ctx, &receipts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	var domainReceipts []Receipt
	for _, receipt := range receipts {
		domainReceipts = append(domainReceipts, *receipt.MapReceipt())
	}

	return domainReceipts, nil
}

func (r *Repository) ListReceiptsPreviousMonth(ctx context.Context, email string) ([]Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "receipt_date").
		From("receipts").
		Where(sq.And{
			sq.Eq{"user_email": email},
			sq.Expr("receipt_date >= date_trunc('month',current_date) - INTERVAL '1' MONTH"),
			sq.Expr("receipt_date < date_trunc('month',current_date)"),
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipts []dbReceipt
	err = r.dbx.SelectContext(ctx, &receipts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	var domainReceipts []Receipt
	for _, receipt := range receipts {
		domainReceipts = append(domainReceipts, *receipt.MapReceipt())
	}

	return domainReceipts, nil
}

func (r *Repository) ListReceiptsPendingReview(ctx context.Context, email string) ([]Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "receipt_date").
		From("receipts").
		Where(sq.And{
			sq.Eq{"user_email": email},
			sq.Eq{"pending_review": true},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipts []dbReceipt
	err = r.dbx.SelectContext(ctx, &receipts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	var domainReceipts []Receipt
	for _, receipt := range receipts {
		domainReceipts = append(domainReceipts, *receipt.MapReceipt())
	}

	return domainReceipts, nil
}

func (r *Repository) GetReceipt(ctx context.Context, receiptID uint64) (*Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "created_at", "receipt_image", "user_email", "receipt_date").
		From("receipts").
		Where(sq.Eq{"id": receiptID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipt dbReceipt
	err = r.dbx.GetContext(ctx, &receipt, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	return receipt.MapReceipt(), nil
}

func (r *Repository) GetReceiptImage(ctx context.Context, receiptID uint64) ([]byte, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("receipt_image").
		From("receipts").
		Where(sq.Eq{"id": receiptID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("compile query: %w", err)
	}

	var receipt dbReceipt
	err = r.dbx.GetContext(ctx, &receipt, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select receipts: %w", err)
	}

	return receipt.Image, nil
}

func (r *Repository) DeleteReceipt(ctx context.Context, id int64) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	txn, err := r.dbx.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer txn.TxClose(ctx)

	query, args, err := psql.Delete("receipts").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.dbx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query to delete receipt: %w", err)
	}

	query, args, err = psql.Delete("expenses").Where(sq.Eq{"receipt_id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.dbx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to execute query to delete expenses: %w", err)
	}

	err = txn.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
