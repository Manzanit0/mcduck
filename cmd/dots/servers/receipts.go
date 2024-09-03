package servers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/jmoiron/sqlx"
	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xtrace"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
)

type receiptsServer struct {
	DB       *sqlx.DB
	Telegram tgram.Client
	Parser   client.ParserClient
	Receipts *receipt.Repository
}

var _ receiptsv1connect.ReceiptsServiceClient = &receiptsServer{}

func NewReceiptsServer(db *sqlx.DB, t tgram.Client) receiptsv1connect.ReceiptsServiceClient {
	return &receiptsServer{DB: db, Telegram: t}
}

func (s *receiptsServer) CreateReceipt(ctx context.Context, req *connect.Request[receiptsv1.CreateReceiptRequest]) (*connect.Response[receiptsv1.CreateReceiptResponse], error) {
	_, span := xtrace.StartSpan(ctx, "Create Receipts")
	defer span.End()

	email := auth.MustGetUserEmailConnect(ctx)

	g, ctx := errgroup.WithContext(ctx)
	for i, file := range req.Msg.ReceiptFiles {
		g.Go(func() error {
			_, span := xtrace.StartSpan(ctx, "Process Receipt")
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

			_, err = s.Receipts.CreateReceipt(ctx, receipt.CreateReceiptRequest{
				Amount:      parsed.Amount,
				Description: parsed.Description,
				Vendor:      parsed.Vendor,
				Image:       file,
				Date:        parsedTime,
				Email:       email,
			})
			if err != nil {
				slog.Error("failed to insert receipt", "error", err.Error(), "index", i)
				return fmt.Errorf("parse receipt: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "create receipt", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&receiptsv1.CreateReceiptResponse{})
	return res, nil
}

func (s *receiptsServer) UpdateReceipt(ctx context.Context, req *connect.Request[receiptsv1.UpdateReceiptRequest]) (*connect.Response[receiptsv1.UpdateReceiptResponse], error) {
	_, span := xtrace.StartSpan(ctx, "Update Receipt")
	defer span.End()

	var date *time.Time
	if req.Msg.Date != nil {
		d, err := time.Parse("2006-01-02", req.Msg.Date.String())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unable to parse date: %w", err))
		}
		date = &d
	}

	dto := receipt.UpdateReceiptRequest{
		ID:            int64(req.Msg.Id),
		Vendor:        req.Msg.Vendor,
		PendingReview: req.Msg.PendingReview,
		Date:          date,
	}

	err := s.Receipts.UpdateReceipt(ctx, dto)
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
