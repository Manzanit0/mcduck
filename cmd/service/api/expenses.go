package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/expense"
)

type ExpensesController struct {
	DB *sqlx.DB
}

type ExpenseViewModel struct {
	ID          string
	Date        string
	Amount      string
	Category    string
	Subcategory string
}

func MapExpenses(expenses []expense.Expense) (models []ExpenseViewModel) {
	for _, e := range expenses {
		models = append(models, ExpenseViewModel{
			ID:          fmt.Sprint(e.ID),
			Date:        e.Date.Format("2006-01-02"),
			Amount:      fmt.Sprintf("%0.2f", e.Amount),
			Category:    strings.Title(e.Category),
			Subcategory: strings.Title(e.Subcategory),
		})
	}

	return
}

func (d *ExpensesController) ListExpenses(c *gin.Context) {
	user := auth.GetUserEmail(c)
	expenses, err := expense.ListExpenses(c.Request.Context(), d.DB, user)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "expenses.html", gin.H{
		"User":        user,
		"HasExpenses": len(expenses) > 0,
		"Expenses":    MapExpenses(expenses),
	})
}

type UpdateExpense struct {
	Date        string  `json:"date,omitempty"`
	Amount      float32 `json:"amount,omitempty,string"`
	Category    string  `json:"category,omitempty"`
	Subcategory string  `json:"subcategory,omitempty"`
}

func (d *ExpensesController) UpdateExpense(c *gin.Context) {
	// TODO: add authorization; any users should not be able to modify any expense

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

	var date time.Time
	if payload.Date != "" {
		date, err = time.Parse("2006-01-02", payload.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse date: %s", err.Error())})
			return
		}
	}

	err = expense.UpdateExpense(c.Request.Context(), d.DB, expense.Expense{
		ID:          i,
		Date:        date,
		Amount:      payload.Amount,
		Category:    payload.Category,
		Subcategory: payload.Subcategory,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to update expense: %s", err.Error())})
		return
	}

	c.JSON(http.StatusAccepted, "")
}
