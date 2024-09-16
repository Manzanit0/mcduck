package servers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"connectrpc.com/connect"
	"github.com/jmoiron/sqlx"
	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xtrace"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type receiptsServer struct {
	Telegram tgram.Client
	Parser   client.ParserClient
	Receipts *receipt.Repository
	Expenses *expense.Repository
}

var _ receiptsv1connect.ReceiptsServiceClient = &receiptsServer{}

func NewReceiptsServer(db *sqlx.DB, p client.ParserClient, t tgram.Client) receiptsv1connect.ReceiptsServiceClient {
	return &receiptsServer{
		Telegram: t,
		Parser:   p,
		Receipts: receipt.NewRepository(db),
		Expenses: expense.NewRepository(db),
	}
}

func (s *receiptsServer) CreateReceipts(ctx context.Context, req *connect.Request[receiptsv1.CreateReceiptsRequest]) (*connect.Response[receiptsv1.CreateReceiptsResponse], error) {
	ctx, span := xtrace.StartSpan(ctx, "Create Receipt")
	defer span.End()

	email := auth.MustGetUserEmailConnect(ctx)

	type receiptWithExpenses struct {
		receipt  *receipt.Receipt
		expenses []expense.Expense
	}

	ch := make(chan receiptWithExpenses, len(req.Msg.ReceiptFiles))

	g, ctx := errgroup.WithContext(ctx)
	for i, file := range req.Msg.ReceiptFiles {
		g.Go(func() error {
			ctx, span := xtrace.StartSpan(ctx, "Process Receipt")
			defer span.End()

			parsed, err := s.Parser.ParseReceipt(ctx, email, file)
			if err != nil {
				slog.Error("failed to parse receipt through parser service", "error", err.Error(), "index", i)
				return fmt.Errorf("parse receipt: %w", err)
			}

			parsedTime, err := time.Parse("02/01/2006", parsed.PurchaseDate)
			if err != nil {
				slog.Info("failed to parse receipt date. Defaulting to 'now' ", "error", err.Error(), "index", i)
				parsedTime = time.Now()
			}

			created, err := s.Receipts.CreateReceipt(ctx, receipt.CreateReceiptRequest{
				Amount:      parsed.Amount,
				Description: parsed.Description,
				Vendor:      parsed.Vendor,
				Image:       file,
				Date:        parsedTime,
				Email:       email,
			})
			if err != nil {
				slog.Error("failed to insert receipt", "error", err.Error(), "index", i)
				return fmt.Errorf("create receipt: %w", err)
			}

			expenses, err := s.Expenses.ListExpensesForReceipt(ctx, uint64(created.ID))
			if err != nil {
				slog.Error("failed to list expenses for receipt", "error", err.Error())
				return fmt.Errorf("list expenses: %w", err)
			}

			ch <- receiptWithExpenses{receipt: created, expenses: expenses}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "create receipt", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	close(ch)

	res := connect.NewResponse(&receiptsv1.CreateReceiptsResponse{})

	for e := range ch {
		res.Msg.Receipts = append(res.Msg.Receipts, &receiptsv1.Receipt{
			Id:       uint64(e.receipt.ID),
			Status:   mapReceiptStatus(e.receipt),
			Vendor:   e.receipt.Vendor,
			Date:     timestamppb.New(e.receipt.Date),
			Expenses: mapExpenses(e.expenses),
		})
	}

	return res, nil
}

func (s *receiptsServer) UpdateReceipt(ctx context.Context, req *connect.Request[receiptsv1.UpdateReceiptRequest]) (*connect.Response[receiptsv1.UpdateReceiptResponse], error) {
	_, span := xtrace.StartSpan(ctx, "Update Receipt")
	defer span.End()

	_, err := s.Receipts.GetReceipt(ctx, req.Msg.Id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("receipt with id %d doesn't exist", req.Msg.Id))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to find receipt: %w", err))
	}

	var date *time.Time
	if req.Msg.Date != nil {
		d := req.Msg.Date.AsTime()
		date = &d
	}

	dto := receipt.UpdateReceiptRequest{
		ID:            int64(req.Msg.Id),
		Vendor:        req.Msg.Vendor,
		PendingReview: req.Msg.PendingReview,
		Date:          date,
	}

	err = s.Receipts.UpdateReceipt(ctx, dto)
	if err != nil {
		slog.Error("failed to update receipt", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to update receipt: %w", err))
	}

	res := connect.NewResponse(&receiptsv1.UpdateReceiptResponse{})
	return res, nil
}

func (s *receiptsServer) DeleteReceipt(ctx context.Context, req *connect.Request[receiptsv1.DeleteReceiptRequest]) (*connect.Response[receiptsv1.DeleteReceiptResponse], error) {
	_, span := xtrace.StartSpan(ctx, "Update Receipt")
	defer span.End()

	err := s.Receipts.DeleteReceipt(ctx, int64(req.Msg.Id))
	if err != nil {
		slog.Error("failed to delete receipt", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to delete receipt: %w", err))
	}

	res := connect.NewResponse(&receiptsv1.DeleteReceiptResponse{})
	return res, nil
}

func (s *receiptsServer) ListReceipts(ctx context.Context, req *connect.Request[receiptsv1.ListReceiptsRequest]) (*connect.Response[receiptsv1.ListReceiptsResponse], error) {
	_, span := xtrace.StartSpan(ctx, "List Receipts")
	defer span.End()

	userEmail := auth.MustGetUserEmailConnect(ctx)

	var receipts []receipt.Receipt
	var err error

	_, span = xtrace.StartSpan(ctx, "Filter Receipts")
	if req.Msg.Since != nil {
		switch *req.Msg.Since {
		case receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_CURRENT_MONTH:
			receipts, err = s.Receipts.ListReceiptsCurrentMonth(ctx, userEmail)
		case receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_PREVIOUS_MONTH:
			receipts, err = s.Receipts.ListReceiptsPreviousMonth(ctx, userEmail)
		case receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_ALL_TIME:
			receipts, err = s.Receipts.ListReceipts(ctx, userEmail)
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported since value"))
		}
	}

	if err != nil {
		slog.Error("failed to list receipts", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to list receipts: %w", err))
	}

	if req.Msg.Status != nil {
		// Note: iterate from the back so we don't have to worry about removed indexes.
		for i := len(receipts) - 1; i >= 0; i-- {
			switch *req.Msg.Status {
			case receiptsv1.ReceiptStatus_RECEIPT_STATUS_PENDING_REVIEW:
				if receipts[i].PendingReview {
					delete(receipts, i)
				}
			case receiptsv1.ReceiptStatus_RECEIPT_STATUS_REVIEWED:
				if !receipts[i].PendingReview {
					delete(receipts, i)
				}
			}
		}
	}
	defer span.End()

	// Sort the most recent first
	_, span = xtrace.StartSpan(ctx, "Sort Receipts")
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].Date.After(receipts[j].Date)
	})
	defer span.End()

	_, span = xtrace.StartSpan(ctx, "Map Receipts to Response")
	defer span.End()

	resReceipts := make([]*receiptsv1.Receipt, len(receipts))
	for i, receipt := range receipts {
		resReceipts[i].Id = uint64(receipt.ID)
		resReceipts[i].Status = mapReceiptStatus(&receipt)
		resReceipts[i].Vendor = receipt.Vendor
		resReceipts[i].Date = timestamppb.New(receipt.Date)

		// FIXME(performance): We should probably do a bulk query before the loop.
		expenses, err := s.Expenses.ListExpensesForReceipt(ctx, uint64(receipt.ID))
		if err != nil {
			slog.Error("failed to list expenses for receipt", "error", err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to list expenses for receipt: %w", err))
		}

		resReceipts[i].Expenses = make([]*receiptsv1.Expense, len(expenses))
		for j, e := range expenses {
			resExp := receiptsv1.Expense{
				Id:          e.ID,
				Date:        timestamppb.New(e.Date),
				Category:    e.Category,
				Subcategory: e.Subcategory,
				Description: e.Description,
				Amount:      uint64(expense.ConvertToCents(e.Amount)),
			}

			resReceipts[i].Expenses[j] = &resExp
		}

	}

	res := connect.NewResponse(&receiptsv1.ListReceiptsResponse{Receipts: resReceipts})
	return res, nil
}

func (s *receiptsServer) GetReceipt(ctx context.Context, req *connect.Request[receiptsv1.GetReceiptRequest]) (*connect.Response[receiptsv1.GetReceiptResponse], error) {
	_, span := xtrace.StartSpan(ctx, "Get Receipt")
	defer span.End()

	receipt, err := s.Receipts.GetReceipt(ctx, req.Msg.Id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("receipt with id %d doesn't exist", req.Msg.Id))
	} else if err != nil {
		slog.Error("failed to get receipt", "error", err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get receipt: %w", err))
	}

	expenses, err := s.Expenses.ListExpensesForReceipt(ctx, req.Msg.Id)
	if err != nil {
		slog.Error("failed to list expenses for receipt", "error", err.Error())
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get expenses for receipt: %w", err))
	}

	res := connect.NewResponse(&receiptsv1.GetReceiptResponse{
		Receipt: &receiptsv1.FullReceipt{
			Id:       uint64(receipt.ID),
			Status:   mapReceiptStatus(receipt),
			Vendor:   receipt.Vendor,
			Date:     timestamppb.New(receipt.Date),
			File:     receipt.Image,
			Expenses: mapExpenses(expenses),
		},
	})

	return res, nil
}

func mapExpenses(expenses []expense.Expense) []*receiptsv1.Expense {
	resExpenses := make([]*receiptsv1.Expense, len(expenses))
	for i, e := range expenses {
		resExp := receiptsv1.Expense{
			Id:          e.ID,
			Date:        timestamppb.New(e.Date),
			Category:    e.Category,
			Subcategory: e.Subcategory,
			Description: e.Description,
			Amount:      uint64(expense.ConvertToCents(e.Amount)),
		}

		resExpenses[i] = &resExp
	}

	return resExpenses
}

func delete[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func mapReceiptStatus(r *receipt.Receipt) receiptsv1.ReceiptStatus {
	if r.PendingReview {
		return receiptsv1.ReceiptStatus_RECEIPT_STATUS_PENDING_REVIEW
	}

	return receiptsv1.ReceiptStatus_RECEIPT_STATUS_REVIEWED
}
