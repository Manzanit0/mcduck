package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/pkg/auth"
)

type ExpensesController struct {
	Expenses *expense.Repository
}

type ExpenseViewModel struct {
	ID          string
	Date        string
	Amount      string
	Category    string
	Subcategory string
	Description string
	ReceiptID   uint64
}

func MapExpenses(expenses []expense.Expense) (models []ExpenseViewModel) {
	for _, e := range expenses {
		models = append(models, ExpenseViewModel{
			ID:          fmt.Sprint(e.ID),
			Date:        e.Date.Format("2006-01-02"),
			Amount:      fmt.Sprintf("%0.2f", e.Amount),
			Category:    strings.Title(e.Category),
			Subcategory: strings.Title(e.Subcategory),
			Description: e.Description,
			ReceiptID:   e.ReceiptID,
		})
	}

	return
}

func (d *ExpensesController) ListExpenses(c *gin.Context) {
	user := auth.GetUserEmail(c)
	expenses, err := d.Expenses.ListExpenses(c.Request.Context(), user)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Sort the most recent first
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Date.After(expenses[j].Date)
	})

	c.HTML(http.StatusOK, "expenses.html", gin.H{
		"User":        user,
		"HasExpenses": len(expenses) > 0,
		"Expenses":    MapExpenses(expenses),
	})
}

type UpdateExpense struct {
	Date        *string  `json:"date"`
	Amount      *float32 `json:"amount,string"`
	Category    *string  `json:"category"`
	Subcategory *string  `json:"subcategory"`
	Description *string  `json:"description"`
	ReceiptID   *uint64  `json:"receipt_id,string"`
}

func (d *ExpensesController) UpdateExpense(c *gin.Context) {
	payload := UpdateExpense{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse request body: %s", err.Error())})
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
		return
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

	err = d.Expenses.UpdateExpense(c.Request.Context(), expense.UpdateExpenseRequest{
		ID:          i,
		Date:        date,
		Amount:      payload.Amount,
		Category:    payload.Category,
		Subcategory: payload.Subcategory,
		Description: payload.Description,
		ReceiptID:   payload.ReceiptID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to update expense: %s", err.Error())})
		return
	}

	c.JSON(http.StatusAccepted, "")
}

type CreateExpensePayload struct {
	Date      string  `json:"date"`
	Amount    float32 `json:"amount,string"`
	ReceiptID *uint64 `json:"receipt_id,string"`
}

type CreateExpenseResponse struct {
	ID int64 `json:"id"`
}

func (d *ExpensesController) CreateExpense(c *gin.Context) {
	payload := CreateExpensePayload{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse request body: %s", err.Error())})
		return
	}

	date, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse date: %s", err.Error())})
		return
	}

	expenseID, err := d.Expenses.CreateExpense(c.Request.Context(), expense.CreateExpenseRequest{
		UserEmail: auth.GetUserEmail(c),
		Date:      date,
		Amount:    payload.Amount,
		ReceiptID: payload.ReceiptID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create expense: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, CreateExpenseResponse{ID: expenseID})
}

func (d *ExpensesController) DeleteExpense(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
		return
	}

	err = d.Expenses.DeleteExpense(c.Request.Context(), i)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to delete expense: %s", err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, "")
}

type MergeExpensesPayload struct {
	ReceiptID  uint64  `json:"receipt_id,string"`
	ExpenseIDs []int64 `json:"expense_ids"`
}

func (d *ExpensesController) MergeExpenses(c *gin.Context) {
	payload := MergeExpensesPayload{}
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse request body: %s", err.Error())})
		return
	}

	var total float32
	expenses := []*expense.Expense{}
	for _, id := range payload.ExpenseIDs {
		expense, err := d.Expenses.FindExpense(c.Request.Context(), int64(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to fetch expense: %s", err.Error())})
			return
		}

		if expense.ReceiptID != payload.ReceiptID {
			errorMessage := fmt.Sprintf("expense with ID %d doesn't belong to receipt %d", expense.ID, payload.ReceiptID)
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
			return
		}

		// FIXME: this operation should not be done with floats.
		total += expense.Amount
		expenses = append(expenses, expense)
	}

	// FIXME: create and delete should be done atomically
	expenseID, err := d.Expenses.CreateExpense(c.Request.Context(), expense.CreateExpenseRequest{
		UserEmail: auth.GetUserEmail(c),
		Date:      time.Now(),
		Amount:    total,
		ReceiptID: &payload.ReceiptID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create expense: %s", err.Error())})
		return
	}

	for _, id := range payload.ExpenseIDs {
		err = d.Expenses.DeleteExpense(c.Request.Context(), int64(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to delete expense: %s", err.Error())})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"expense_id": expenseID})
}
