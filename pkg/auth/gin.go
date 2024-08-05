package auth

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authCookieName string = "_mcduck_key"
	userContextKey string = "user.email"
)

func CookieMiddleware(c *gin.Context) {
	if _, exists := c.Get(userContextKey); exists {
		c.Next()
		return
	}

	token, err := c.Cookie(authCookieName)
	if err != nil {
		c.Next()
		return
	}

	email, isValid := ValidateJWT(token)
	if !isValid {
		c.Next()
		return
	}

	slog.Info("user logged in", "email", email)
	c.Set(userContextKey, email)
	c.Next()
}

func BearerMiddleware(c *gin.Context) {
	if _, exists := c.Get(userContextKey); exists {
		c.Next()
		return
	}

	header := c.GetHeader("authorization")
	s := strings.Split(header, " ")
	if len(s) != 2 {
		c.Next()
		return
	}

	email, isValid := ValidateJWT(s[1])
	if !isValid {
		c.Next()
		return
	}

	slog.Info("user logged in", "email", email)
	c.Set(userContextKey, email)
	c.Next()
}

func GetUserEmail(c *gin.Context) string {
	return c.GetString(userContextKey)
}

func SetAuthCookie(c *gin.Context, token string) {
	c.SetCookie(authCookieName, token, int(threeDays.Seconds()), "", "", false, true)
}

func RemoveAuthCookie(c *gin.Context) {
	c.SetCookie(authCookieName, "", -1, "", "", false, true)
}
