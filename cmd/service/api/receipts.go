package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/invx"
)

type ReceiptsController struct {
	Expenses *expense.Repository
	Receipts *receipt.Repository
	Invx     invx.Client
}

func (d *ReceiptsController) ListReceipts(c *gin.Context) {
	userEmail := auth.GetUserEmail(c)
	receipts, err := d.Receipts.ListReceipts(c.Request.Context(), userEmail)
	if err != nil {
		log.Println("ListPendingReceipts:", "query receipts:", err.Error())
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
		return
	}

	// Sort the most recent first
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].Date.After(receipts[j].Date)
	})

	viewReceipts := ToReceiptViewModel(receipts)

	// Note: awful stuff. We should probably do this in a single SQL query or something.
	for i, r := range viewReceipts {
		id, err := strconv.ParseUint(r.ID, 10, 64)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
			return
		}

		expenses, err := d.Expenses.ListExpensesForReceipt(c.Request.Context(), id)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": "We were unable to find your receipts - please try again."})
			return
		}

		var total float64
		for _, e := range expenses {
			total += float64(e.Amount)
		}

		viewReceipts[i].TotalAmount = fmt.Sprintf("%0.2f", total)
	}

	c.HTML(http.StatusOK, "list_receipts.html", gin.H{
		"User":        userEmail,
		"HasReceipts": len(receipts) > 0,
		"Receipts":    viewReceipts,
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

	amounts, err := d.Invx.ParseReceipt(c.Request.Context(), data)
	if err != nil {
		log.Println("CreateReceipt:", "invx parse receipt:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to parse receipt: %s", err.Error())})
		return
	}

	email := auth.GetUserEmail(c)
	rcpt, err := d.Receipts.CreateReceipt(c.Request.Context(), data, amounts, email)
	if err != nil {
		log.Println("CreateReceipt:", "insert receipt:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, CreateReceiptResponse{
		ReceiptID: rcpt.ID,
		Amounts:   amounts,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to retrieve receipt: %s", err.Error())})
		return
	}

	expenses, err := d.Expenses.ListExpensesForReceipt(c.Request.Context(), receiptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to list expenses: %s", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "review_receipt.html", gin.H{
		"User":     userEmail,
		"Receipt":  ToSingleReceiptViewModel(receipt),
		"Expenses": MapExpenses(expenses),
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to delete receipt: %s", err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, "")
}

type ReceiptViewModel struct {
	ID            string
	Date          string
	Vendor        string
	PendingReview string
	Image         string
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

	encoded := base64.StdEncoding.EncodeToString(r.Image)

	return ReceiptViewModel{
		ID:            fmt.Sprint(r.ID),
		Date:          r.Date.Format("2006-01-02"),
		Vendor:        strings.Title(r.Vendor),
		PendingReview: pendingReview,
		Image:         encoded,
	}
}
