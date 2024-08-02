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
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
)

type ReceiptsController struct {
	Expenses *expense.Repository
	Receipts *receipt.Repository
	Parser   client.ParserClient
}

func (d *ReceiptsController) ListReceipts(c *gin.Context) {
	var receipts []receipt.Receipt
	var err error

	userEmail := auth.GetUserEmail(c)
	switch c.Query("when") {
	case "current_month":
		receipts, err = d.Receipts.ListReceiptsCurrentMonth(c.Request.Context(), userEmail)
	case "previous_month":
		receipts, err = d.Receipts.ListReceiptsPreviousMonth(c.Request.Context(), userEmail)
	case "all_time", "":
		receipts, err = d.Receipts.ListReceipts(c.Request.Context(), userEmail)
	default:
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Unsupported query value for 'when'"})
		return
	}

	receiptStatus := c.Query("status")
	if receiptStatus == "pending_review" {
		receipts, err = d.Receipts.ListReceiptsPendingReview(c.Request.Context(), userEmail)
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

		expenses, err := d.Expenses.ListExpensesForReceipt(c.Request.Context(), id)
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

type CreateReceiptResponse struct {
	ReceiptID int64              `json:"receipt_id"`
	Amounts   map[string]float64 `json:"receipt_amounts"`
}

func (d *ReceiptsController) CreateReceipt(c *gin.Context) {
	file, err := c.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to read file: %s", err.Error())})
		return
	}

	dir := os.TempDir()
	filePath := filepath.Join(dir, file.Filename)
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to save file to disk: %s", err.Error())})
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to read file from disk: %s", err.Error())})
		return
	}

	email := auth.GetUserEmail(c)

	parsed, err := d.Parser.ParseReceipt(c.Request.Context(), email, data)
	if err != nil {
		slog.Error("failed to parse receipt through parser service", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to parse receipt: %s", err.Error())})
		return
	}

	parsedTime, err := time.Parse("02/01/2006", parsed.PurchaseDate)
	if err != nil {
		slog.Info("failed to parse receipt date. Defaulting to 'now' ", "error", err.Error())
		parsedTime = time.Now()
	}

	rcpt, err := d.Receipts.CreateReceipt(c.Request.Context(), receipt.CreateReceiptRequest{
		Amount:      parsed.Amount,
		Description: parsed.Description,
		Vendor:      parsed.Vendor,
		Image:       data,
		Date:        parsedTime,
		Email:       email,
	})
	if err != nil {
		slog.Error("failed to insert receipt", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, CreateReceiptResponse{
		ReceiptID: rcpt.ID,
		Amounts:   map[string]float64{parsed.Description: parsed.Amount},
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

	g, ctx := errgroup.WithContext(ctx)
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

			email := auth.GetUserEmail(c)
			parsed, err := d.Parser.ParseReceipt(ctx, email, data)
			if err != nil {
				return fmt.Errorf("parse receipt: %w", err)
			}

			parsedTime, err := time.Parse("02/01/2006", parsed.PurchaseDate)
			if err != nil {
				slog.Info("failed to parse receipt date. Defaulting to 'now' ", "error", err.Error())
				parsedTime = time.Now()
			}

			_, err = d.Receipts.CreateReceipt(ctx, receipt.CreateReceiptRequest{
				Amount:      parsed.Amount,
				Description: parsed.Description,
				Vendor:      parsed.Vendor,
				Image:       data,
				Date:        parsedTime,
				Email:       email,
			})
			if err != nil {
				return fmt.Errorf("create receipt: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
