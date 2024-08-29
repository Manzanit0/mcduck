package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	authv1 "github.com/manzanit0/mcduck/api/auth.v1"
	"github.com/manzanit0/mcduck/api/auth.v1/authv1connect"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
)

type RegistrationController struct {
	DB              *sqlx.DB
	Telegram        tgram.Client
	AuthServiceHost string
	AuthClient      authv1connect.AuthServiceClient
}

func (r *RegistrationController) GetRegisterForm(c *gin.Context) {
	url := fmt.Sprintf("%s%s", r.AuthServiceHost, authv1connect.AuthServiceRegisterProcedure)
	c.HTML(http.StatusOK, "register.html", gin.H{"RegisterEndpointURL": url})
}

func (r *RegistrationController) GetLoginForm(c *gin.Context) {
	url := fmt.Sprintf("%s%s", r.AuthServiceHost, authv1connect.AuthServiceLoginProcedure)
	c.HTML(http.StatusOK, "login.html", gin.H{"LoginEndpointURL": url})
}

func (_ *RegistrationController) Signout(c *gin.Context) {
	if email := auth.GetUserEmail(c); email != "" {
		auth.RemoveAuthCookie(c)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{})
}

// GetConnectForm returns the HTML form to connect a telegram account to a
// mcduck account.
func (r *RegistrationController) GetConnectForm(c *gin.Context) {
	idStr := c.Query("tgram")
	if idStr == "" {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "Are you trying to connect our Telegram bot with the web app? Please ask the bot for another link, this one seems funny."})
		return
	}

	// If the user is logged in, no point in showing the login form again.
	if email := auth.GetUserEmail(c); email != "" {
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		_, err = r.AuthClient.ConnectTelegram(c.Request.Context(), connect.NewRequest(&authv1.ConnectTelegramRequest{
			Email:  email,
			ChatId: idInt,
		}))
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
		return
	}

	url := fmt.Sprintf("%s%s", r.AuthServiceHost, authv1connect.AuthServiceConnectTelegramProcedure)
	c.HTML(http.StatusOK, "telegram_connect.html", gin.H{"TelegramChatID": idStr, "ConnectEndpointURL": url})
}
