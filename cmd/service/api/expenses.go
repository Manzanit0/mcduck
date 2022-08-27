package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/expense"
)

type ExpensesController struct {
	DB *sqlx.DB
}

type ExpenseViewModel struct {
	Date        string
	Amount      string
	Category    string
	Subcategory string
}

func MapExpenses(expenses []expense.Expense) (models []ExpenseViewModel) {
	for _, e := range expenses {
		models = append(models, ExpenseViewModel{
			Date:        e.Date.Format("2006-01-02"),
			Amount:      fmt.Sprintf("%0.2f€", e.Amount),
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
