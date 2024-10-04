package servers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jmoiron/sqlx"
	expensesv1 "github.com/manzanit0/mcduck/api/expenses.v1"
	"github.com/manzanit0/mcduck/api/expenses.v1/expensesv1connect"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/pkg/xtrace"
)

type expensesServer struct {
	Expenses *expense.Repository
}

var _ expensesv1connect.ExpensesServiceClient = &expensesServer{}

func NewExpensesServer(db *sqlx.DB) *expensesServer {
	return &expensesServer{
		Expenses: expense.NewRepository(db),
	}
}

// CreateExpense implements expensesv1connect.ExpensesServiceClient.
func (e *expensesServer) CreateExpense(context.Context, *connect.Request[expensesv1.CreateExpenseRequest]) (*connect.Response[expensesv1.CreateExpenseResponse], error) {
	panic("unimplemented")
}

// DeleteExpense implements expensesv1connect.ExpensesServiceClient.
func (e *expensesServer) DeleteExpense(context.Context, *connect.Request[expensesv1.DeleteExpenseRequest]) (*connect.Response[expensesv1.DeleteExpenseResponse], error) {
	panic("unimplemented")
}

// ListExpenses implements expensesv1connect.ExpensesServiceClient.
func (e *expensesServer) ListExpenses(context.Context, *connect.Request[expensesv1.ListExpensesRequest]) (*connect.Response[expensesv1.ListExpensesResponse], error) {
	panic("unimplemented")
}

type UpdateExpense struct {
	Date        *string  `json:"date"`
	Amount      *float32 `json:"amount,string"`
	Category    *string  `json:"category"`
	Subcategory *string  `json:"subcategory"`
	Description *string  `json:"description"`
	ReceiptID   *uint64  `json:"receipt_id,string"`
}

// UpdateExpense implements expensesv1connect.ExpensesServiceClient.
func (e *expensesServer) UpdateExpense(ctx context.Context, req *connect.Request[expensesv1.UpdateExpenseRequest]) (*connect.Response[expensesv1.UpdateExpenseResponse], error) {
	ctx, span := xtrace.GetSpan(ctx)

	_, err := e.Expenses.FindExpense(ctx, int64(req.Msg.Id))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("expense with id %d doesn't exist", req.Msg.Id))
	} else if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to find expense: %w", err))
	}

	var date *time.Time
	if req.Msg.Date != nil {
		d := req.Msg.Date.AsTime()
		date = &d
	}

	var amount *float32
	if req.Msg.Amount != nil {
		a := expense.ConvertToDollar(int32(*req.Msg.Amount))
		amount = &a
	}

	err = e.Expenses.UpdateExpense(ctx, expense.UpdateExpenseRequest{
		ID:          int64(req.Msg.Id),
		Date:        date,
		Amount:      amount,
		Category:    req.Msg.Category,
		Subcategory: req.Msg.Subcategory,
		Description: req.Msg.Description,
		ReceiptID:   req.Msg.ReceiptId,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update expense", "error", err.Error())
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to update expense: %w", err))
	}

	// TODO: merge this with the UpdateExpense SQL call via RETURNING.
	exp, err := e.Expenses.FindExpense(ctx, int64(req.Msg.Id))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to find expense: %w", err))
	}

	var receiptID *uint64
	if exp.ReceiptID != 0 {
		receiptID = &exp.ReceiptID
	}

	res := connect.NewResponse(&expensesv1.UpdateExpenseResponse{
		Expense: &expensesv1.Expense{
			Id:          exp.ID,
			ReceiptId:   receiptID,
			Amount:      uint64(expense.ConvertToCents(exp.Amount)),
			Date:        timestamppb.New(exp.Date),
			Category:    exp.Category,
			Subcategory: exp.Subcategory,
			Description: exp.Description,
		},
	})

	return res, nil
}
