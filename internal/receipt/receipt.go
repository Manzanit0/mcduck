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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type dbReceipt struct {
	ID            int64   `db:"id"`
	PendingReview bool    `db:"pending_review"`
	Image         []byte  `db:"receipt_image"`
	UserEmail     string  `db:"user_email"`
	Vendor        *string `db:"vendor"`

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

func (r *Repository) CreateReceipt(ctx context.Context, receiptImage []byte, amounts map[string]float64, userEmail string) (*Receipt, error) {
	if len(receiptImage) == 0 {
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
		Columns("receipt_image", "pending_review", "user_email").
		Values(receiptImage, true, userEmail).
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

	if len(amounts) > 0 {
		e := expense.ExpensesBatch{UserEmail: userEmail}
		for item, amount := range amounts {
			e.Records = append(e.Records, expense.Expense{
				ReceiptID:   uint64(record.ID),
				Date:        time.Now(),
				Amount:      float32(amount),
				UserEmail:   userEmail,
				Description: item,
				Category:    "Receipt Upload",
			})
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
}

func (r *Repository) UpdateReceipt(ctx context.Context, e UpdateReceiptRequest) error {
	var shouldUpdate bool

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

	if !shouldUpdate {
		return nil
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("compile query: %w", err)
	}

	_, err = r.dbx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

func (r *Repository) ListReceipts(ctx context.Context, email string) ([]Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "created_at", "receipt_image").
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

func (r *Repository) GetReceipt(ctx context.Context, receiptID uint64) (*Receipt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.
		Select("id", "vendor", "pending_review", "created_at", "receipt_image", "user_email").
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
