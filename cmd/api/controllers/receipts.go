package controllers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
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
	ctx, span := xtrace.StartSpan(c.Request.Context(), "List Receipts Page")
	defer span.End()

	var receipts []receipt.Receipt
	var err error

	userEmail := auth.GetUserEmail(c)
	switch c.Query("when") {
	case "current_month":
		receipts, err = d.Receipts.ListReceiptsCurrentMonth(ctx, userEmail)
	case "previous_month":
		receipts, err = d.Receipts.ListReceiptsPreviousMonth(ctx, userEmail)
	case "all_time", "":
		receipts, err = d.Receipts.ListReceipts(ctx, userEmail)
	default:
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unsupported query value for 'when'"})
		return
	}

	receiptStatus := c.Query("status")
	if receiptStatus == "pending_review" {
		receipts, err = d.Receipts.ListReceiptsPendingReview(ctx, userEmail)
	}

	if err != nil {
		slog.Error("failed to list receipts", "error", err.Error())
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
		return
	}

	// Sort the most recent first
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].Date.After(receipts[j].Date)
	})

	viewReceipts := ToReceiptViewModel(receipts)

	var pendingReview []ReceiptViewModel
	var reviewed []ReceiptViewModel

	// Note: awful stuff. We should probably do this in a single SQL query or something.
	for i, r := range viewReceipts {
		id, err := strconv.ParseUint(r.ID, 10, 64)
		if err != nil {
			slog.Error("failed to parse receipt ID", "error", err.Error())
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
			return
		}

		expenses, err := d.Expenses.ListExpensesForReceipt(ctx, id)
		if err != nil {
			slog.Error("failed to list expenses for receipt", "error", err.Error())
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
			return
		}

		var total float64
		for _, e := range expenses {
			total += float64(e.Amount)
		}

		viewReceipts[i].TotalAmount = fmt.Sprintf("%0.2f", total)

		if r.PendingReview == "Yes" {
			pendingReview = append(pendingReview, r)
		} else {
			reviewed = append(reviewed, r)
		}
	}

	c.HTML(http.StatusOK, "list_receipts.html", gin.H{
		"User":                  userEmail,
		"HasReceipts":           len(receipts) > 0,
		"Receipts":              viewReceipts,
		"ReceiptsPendingReview": len(pendingReview),
		"ReceiptsReviewed":      len(reviewed),
	})
}

type UpdateReceiptRequest struct {
	Vendor        *string `json:"vendor"`
	PendingReview *string `json:"pending_review"`
	Date          *string `json:"date"`
}

func (d *ReceiptsController) UpdateReceipt(c *gin.Context) {
	payload := UpdateReceiptRequest{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse request body: %s", err.Error())})
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse date: %s", err.Error())})
			return
		}
		date = &d
	}

	err = d.Receipts.UpdateReceipt(c.Request.Context(), receipt.UpdateReceiptRequest{
		ID:            i,
		Vendor:        payload.Vendor,
		PendingReview: pendingReview,
		Date:          date,
	})
	if err != nil {
		slog.Error("failed to update receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to update receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusAccepted, "")
}

func (d *ReceiptsController) ReviewReceipt(c *gin.Context) {
	userEmail := auth.GetUserEmail(c)

	id := c.Param("id")
	receiptID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
		return
	}

	receipt, err := d.Receipts.GetReceipt(c.Request.Context(), receiptID)
	if err != nil {
		slog.Error("failed to get receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to retrieve receipt: %s", err.Error())})
		return
	}

	expenses, err := d.Expenses.ListExpensesForReceipt(c.Request.Context(), receiptID)
	if err != nil {
		slog.Error("failed to list expenses for receipt", "error", err.Error())
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
	id := c.Param("id")
	receiptID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
		return
	}

	image, err := d.Receipts.GetReceiptImage(c.Request.Context(), receiptID)
	if err != nil {
		slog.Error("failed to get receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to retrieve receipt: %s", err.Error())})
		return
	}

	contentType := http.DetectContentType(image)
	c.Writer.Header().Add("Content-Type", contentType)
	c.Writer.Header().Add("Content-Length", strconv.Itoa(len(image)))

	c.Writer.Write(image)
}

func (d *ReceiptsController) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
		return
	}

	err = d.Receipts.DeleteReceipt(c.Request.Context(), i)
	if err != nil {
		slog.Error("failed to delete receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to delete receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, "")
}

func (d *ReceiptsController) UploadReceipts(c *gin.Context) {
	ctx := c.Request.Context()
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
		slog.Error("failed to copy auth header", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	_, err = d.ReceiptsClient.CreateReceipts(ctx, &req)
	if err != nil {
		slog.Error("failed to create receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	d.ListReceipts(c)
}

type ReceiptViewModel struct {
	ID            string
	Date          string
	Vendor        string
	PendingReview string
	IsPDF         bool
	ReceiptID     int
	TotalAmount   string
}

func ToReceiptViewModel(receipts []receipt.Receipt) (models []ReceiptViewModel) {
	for _, r := range receipts {
		models = append(models, ToSingleReceiptViewModel(&r))
	}

	return
}

func ToSingleReceiptViewModel(r *receipt.Receipt) ReceiptViewModel {
	pendingReview := "No"
	if r.PendingReview {
		pendingReview = "Yes"
	}

	isPDF := http.DetectContentType(r.Image) == "application/pdf"

	return ReceiptViewModel{
		ID:            fmt.Sprint(r.ID),
		Date:          r.Date.Format("2006-01-02"),
		Vendor:        strings.Title(r.Vendor),
		PendingReview: pendingReview,
		IsPDF:         isPDF,
	}
}
