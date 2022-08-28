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
	Date        *string  `json:"date"`
	Amount      *float32 `json:"amount,string"`
	Category    *string  `json:"category"`
	Subcategory *string  `json:"subcategory"`
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

	var date *time.Time
	if payload.Date != nil {
		d, err := time.Parse("2006-01-02", *payload.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse date: %s", err.Error())})
			return
		}
		date = &d
	}

	err = expense.UpdateExpense(c.Request.Context(), d.DB, expense.UpdateExpenseRequest{
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

// ExpenseOwnershipWall validates that the expense ID in the URL parameter
// belongs to the requesting user, otherwise abouts with Unauthorised status.
func ExpenseOwnershipWall(db *sqlx.DB, controller gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
			return
		}

		e, err := expense.FindExpense(c.Request.Context(), db, i)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to find expense: %s", err.Error())})
			return
		}

		if !strings.EqualFold(e.UserEmail, auth.GetUserEmail(c)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "the expense doesn't belong to requesting user"})
			return
		}

		c.Next()
		controller(c)
	}
}
