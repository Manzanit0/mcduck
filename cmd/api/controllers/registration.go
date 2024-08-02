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
	"github.com/manzanit0/mcduck/pkg/xhttp"
)

type RegistrationController struct {
	DB              *sqlx.DB
	Telegram        tgram.Client
	AuthServiceHost string
}

func (r *RegistrationController) GetRegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"RegisterEndpointURL": fmt.Sprintf("%s/auth.v1.AuthService/Register", r.AuthServiceHost),
	})
}

func (r *RegistrationController) GetLoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"LoginEndpointURL": fmt.Sprintf("%s/auth.v1.AuthService/Login", r.AuthServiceHost),
	})
}

// GetConnectForm returns the HTML form to connect a telegram account to a
// mcduck account.
func (r *RegistrationController) GetConnectForm(c *gin.Context) {
	id := c.Query("tgram")
	if id == "" {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "Are you trying to connect our Telegram bot with the web app? Please ask the bot for another link, this one seems funny."})
		return
	}

	// If the user is logged in, no point in showing the login form again.
	if email := auth.GetUserEmail(c); email != "" {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		client := authv1connect.NewAuthServiceClient(xhttp.NewClient(), r.AuthServiceHost)
		_, err = client.ConnectTelegram(c.Request.Context(), connect.NewRequest(&authv1.ConnectTelegramRequest{
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

	c.HTML(http.StatusOK, "telegram_connect.html", gin.H{
		"TelegramChatID":     id,
		"ConnectEndpointURL": fmt.Sprintf("%s%s", r.AuthServiceHost, authv1connect.AuthServiceConnectTelegramProcedure),
	})
}

func (_ *RegistrationController) Signout(c *gin.Context) {
	if email := auth.GetUserEmail(c); email != "" {
		auth.RemoveAuthCookie(c)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{})
}
