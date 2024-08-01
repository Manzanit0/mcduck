package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/manzanit0/mcduck/internal/users"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
)

type UserPayload struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

type ConnectTelegramPayload struct {
	Email          string `form:"email"`
	Password       string `form:"password"`
	TelegramChatID int64  `form:"telegram_chat_id"`
}

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
	}

	// If the user is logged in, no point in showing the login form again.
	if email := auth.GetUserEmail(c); email != "" {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		err = doConnect(c, r.DB, r.Telegram, &users.User{Email: email}, idInt)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
	}

	c.HTML(http.StatusOK, "telegram_connect.html", gin.H{"TelegramChatID": id})
}

// ConnectUser logs in and connects the user to their telegram account.
func (r *RegistrationController) ConnectUser(c *gin.Context) {
	payload := ConnectTelegramPayload{}
	err := c.ShouldBind(&payload)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	user, err := doLogin(c, r.DB, payload.Email, payload.Password)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	err = doConnect(c, r.DB, r.Telegram, user, payload.TelegramChatID)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{"User": user.Email})
}

func doConnect(c *gin.Context, db *sqlx.DB, tgramClient tgram.Client, user *users.User, chatID int64) error {
	err := users.UpdateTelegramChatID(c.Request.Context(), db, user, chatID)
	if err != nil {
		return err
	}

	err = tgramClient.SendMessage(tgram.SendMessageRequest{
		ChatID:    chatID,
		Text:      "You account has been successfully linked\\!",
		ParseMode: tgram.ParseModeMarkdownV2,
	})
	if err != nil {
		return fmt.Errorf("Your account has been linked successfully but we were unable to notify you via Telegram: %w", err)
	}

	return nil
}

func doLogin(c *gin.Context, db *sqlx.DB, email, password string) (*users.User, error) {
	user, err := users.Find(c.Request.Context(), db, email)
	if err != nil {
		slog.Error("unable to find user", "email", email, "error", err.Error())
		return nil, fmt.Errorf("invalid email or password")
	}

	if !auth.CheckPasswordHash(password, user.HashedPassword) {
		slog.Error("invalid password", "email", email, "error", "hashed password doesn't match")
		return nil, fmt.Errorf("invalid email or password")
	}

	err = setCookieAuth(c, email)
	if err != nil {
		return nil, fmt.Errorf("set cooking auth: %w", err)
	}

	return user, nil
}

func (_ *RegistrationController) Signout(c *gin.Context) {
	if email := auth.GetUserEmail(c); email != "" {
		auth.RemoveAuthCookie(c)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func setCookieAuth(c *gin.Context, email string) error {
	token, err := auth.GenerateJWT(email)
	if err != nil {
		return fmt.Errorf("unable to generate JWT: %w", err)
	}

	auth.SetAuthCookie(c, token)
	return nil
}
