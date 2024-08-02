package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
)

// ForceLogin validates that the user is logged in. If they are not it redirects
// them to the login page instead of continuing to the requested page.
func ForceLogin(c *gin.Context) {
	user := auth.GetUserEmail(c)
	if user == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
	c.Next()
}

func ForceAuthentication(c *gin.Context) {
	user := auth.GetUserEmail(c)
	if user == "" {
		c.JSON(http.StatusUnauthorized, gin.H{})
		c.Abort()
		return
	}

	c.Next()
}

// ExpenseOwnershipWall validates that the expense ID in the URL parameter
// belongs to the requesting user, otherwise abouts with Unauthorised status.
func ExpenseOwnershipWall(repo *expense.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse expense id: %s", err.Error())})
			return
		}

		e, err := repo.FindExpense(c.Request.Context(), i)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to find expense: %s", err.Error())})
			return
		}

		if !strings.EqualFold(e.UserEmail, auth.GetUserEmail(c)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "the expense doesn't belong to requesting user"})
			return
		}

		c.Next()
	}
}

func ReceiptOwnershipWall(repo *receipt.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		i, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to parse receipt id: %s", err.Error())})
			return
		}

		receipt, err := repo.GetReceipt(c.Request.Context(), i)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to find receipt: %s", err.Error())})
			return
		}

		if !strings.EqualFold(receipt.UserEmail, auth.GetUserEmail(c)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "the receipt doesn't belong to requesting user"})
			return
		}

		c.Next()
	}
}
