package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/mcduck/internal/users"
)

type UsersController struct {
	DB *sqlx.DB
}

type UserResult struct {
	Email          string `json:"email"`
	TelegramChatID *int64 `json:"telegram_chat_id"`
}

// GET /api/users?chat_id=...
func (d *UsersController) SearchUser(c *gin.Context) {
	chatID := c.Query("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search is only possible by chat_id"})
		return
	}

	u, err := users.FindByChatID(c.Request.Context(), d.DB, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if u == nil {
		c.JSON(http.StatusOK, gin.H{"user": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": UserResult{
		Email:          u.Email,
		TelegramChatID: u.TelegramChatID,
	}})
}
