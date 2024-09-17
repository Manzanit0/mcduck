package controllers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/xtrace"
)

type ReceiptsController struct {
	Expenses *expense.Repository
	Receipts *receipt.Repository
	Parser   client.ParserClient

	ReceiptsClient receiptsv1connect.ReceiptsServiceClient
}

func (d *ReceiptsController) ListReceipts(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	var since receiptsv1.ListReceiptsSince
	switch c.Query("when") {
	case "current_month":
		since = receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_CURRENT_MONTH
	case "previous_month":
		since = receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_PREVIOUS_MONTH
	case "all_time", "":
		since = receiptsv1.ListReceiptsSince_LIST_RECEIPTS_SINCE_ALL_TIME
	}

	var status receiptsv1.ReceiptStatus
	switch c.Query("status") {
	case "pending_review":
		status = receiptsv1.ReceiptStatus_RECEIPT_STATUS_PENDING_REVIEW
	case "reviewed":
		status = receiptsv1.ReceiptStatus_RECEIPT_STATUS_REVIEWED
	}

	req := connect.Request[receiptsv1.ListReceiptsRequest]{
		Msg: &receiptsv1.ListReceiptsRequest{
			Since:  since,
			Status: status,
		},
	}

	err := auth.CopyAuthHeader(&req, c.Request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to copy auth header", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to list receipts: %s", err.Error())})
		return
	}

	res, err := d.ReceiptsClient.ListReceipts(ctx, &req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to list receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to list receipt: %s", err.Error())})
		return
	}

	_, span = xtrace.StartSpan(ctx, "Map view models")
	defer span.End()

	var pendingReviewCount int
	var reviewedCount int
	var viewModels []ReceiptViewModel
	for _, r := range res.Msg.Receipts {
		pendingReview := "No"
		if r.Status == receiptsv1.ReceiptStatus_RECEIPT_STATUS_PENDING_REVIEW {
			pendingReview = "Yes"
			pendingReviewCount += 1
		} else {
			reviewedCount += 1
		}

		var total float64
		for _, expense := range r.Expenses {
			total += float64(expense.Amount)
		}

		v := ReceiptViewModel{
			ID:            fmt.Sprint(r.Id),
			Date:          r.Date.AsTime().Format("2006-01-02"),
			Vendor:        strings.Title(r.Vendor),
			PendingReview: pendingReview,
			TotalAmount:   fmt.Sprintf("%0.2f", total),
		}

		viewModels = append(viewModels, v)
	}

	c.HTML(http.StatusOK, "list_receipts.html", gin.H{
		"User":                  auth.GetUserEmail(c),
		"HasReceipts":           len(viewModels) > 0,
		"Receipts":              viewModels,
		"ReceiptsPendingReview": pendingReviewCount,
		"ReceiptsReviewed":      reviewedCount,
	})
}

type UpdateReceiptRequest struct {
	Vendor        *string `json:"vendor"`
	PendingReview *string `json:"pending_review"`
	Date          *string `json:"date"`
}

func (d *ReceiptsController) UpdateReceipt(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	payload := UpdateReceiptRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse request body: %s", err.Error())})
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
		return
	}

	var pendingReview *bool
	if payload.PendingReview != nil && *payload.PendingReview == "Yes" {
		b := true
		pendingReview = &b
	} else if payload.PendingReview != nil && *payload.PendingReview == "No" {
		b := false
		pendingReview = &b
	}

	var date *time.Time
	if payload.Date != nil {
		d, err := time.Parse("2006-01-02", *payload.Date)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse date: %s", err.Error())})
			return
		}
		date = &d
	}

	err = d.Receipts.UpdateReceipt(ctx, receipt.UpdateReceiptRequest{
		ID:            i,
		Vendor:        payload.Vendor,
		PendingReview: pendingReview,
		Date:          date,
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to update receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to update receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusAccepted, "")
}

func (d *ReceiptsController) ReviewReceipt(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	userEmail := auth.GetUserEmail(c)

	id := c.Param("id")
	receiptID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
		return
	}

	receipt, err := d.Receipts.GetReceipt(ctx, receiptID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to get receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to retrieve receipt: %s", err.Error())})
		return
	}

	expenses, err := d.Expenses.ListExpensesForReceipt(ctx, receiptID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to list expenses for receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to list expenses: %s", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "review_receipt.html", gin.H{
		"User":     userEmail,
		"Receipt":  ToSingleReceiptViewModel(receipt),
		"Expenses": MapExpenses(expenses),
	})
}

func (d *ReceiptsController) GetImage(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	id := c.Param("id")
	receiptID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
		return
	}

	image, err := d.Receipts.GetReceiptImage(ctx, receiptID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to get receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to retrieve receipt: %s", err.Error())})
		return
	}

	contentType := http.DetectContentType(image)
	c.Writer.Header().Add("Content-Type", contentType)
	c.Writer.Header().Add("Content-Length", strconv.Itoa(len(image)))

	c.Writer.Write(image)
}

func (d *ReceiptsController) DeleteReceipt(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
		return
	}

	err = d.Receipts.DeleteReceipt(ctx, i)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to delete receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to delete receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, "")
}

func (d *ReceiptsController) UploadReceipts(c *gin.Context) {
	ctx, span := xtrace.GetSpan(c.Request.Context())

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "get form error: %s", err.Error())
		return
	}

	if len(form.File["files"]) == 0 {
		c.String(http.StatusBadRequest, "no files uploaded")
		return
	}

	var files [][]byte
	var m sync.Mutex
	g, _ := errgroup.WithContext(ctx)
	for _, file := range form.File["files"] {
		g.Go(func() error {
			filename := filepath.Base(file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				return fmt.Errorf("save uploaded file: %w", err)
			}

			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			m.Lock()
			files = append(files, data)
			m.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to read file", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := connect.Request[receiptsv1.CreateReceiptsRequest]{
		Msg: &receiptsv1.CreateReceiptsRequest{
			ReceiptFiles: files,
		},
	}

	err = auth.CopyAuthHeader(&req, c.Request)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to copy auth header", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	_, err = d.ReceiptsClient.CreateReceipts(ctx, &req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "failed to create receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	// NOTE: Since most controller functions rely on an existing span, starting a
	// span here helps group all subspans.
	_, span = xtrace.StartSpan(ctx, "List Receipts Page")
	defer span.End()

	d.ListReceipts(c)
}

type ReceiptViewModel struct {
	ID            string
	Date          string
	Vendor        string
	PendingReview string
	ReceiptID     int
	TotalAmount   string
}

func ToSingleReceiptViewModel(r *receipt.Receipt) ReceiptViewModel {
	pendingReview := "No"
	if r.PendingReview {
		pendingReview = "Yes"
	}

	return ReceiptViewModel{
		ID:            fmt.Sprint(r.ID),
		Date:          r.Date.Format("2006-01-02"),
		Vendor:        strings.Title(r.Vendor),
		PendingReview: pendingReview,
	}
}
