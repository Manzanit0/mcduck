package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/users"
)

const authCookieName string = "_mcduck_key"
const userContextKey string = "user.email"

type UserPayload struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func GetRegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func RegisterUser(c *gin.Context) {
	payload := UserPayload{}
	err := c.ShouldBind(&payload)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	db, ok := c.Get("db")
	if !ok {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	dbx := db.(*sqlx.DB)

	_, err = users.Create(c.Request.Context(), dbx, users.User{Email: payload.Email, Password: payload.Password})
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateJWT(payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(authCookieName, token, 3600, "", "", false, true)

	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func GetLoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func LoginUser(c *gin.Context) {
	payload := UserPayload{}
	err := c.ShouldBind(&payload)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	db, ok := c.Get("db")
	if !ok {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	dbx := db.(*sqlx.DB)

	user, err := users.Find(c.Request.Context(), dbx, payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	if !auth.CheckPasswordHash(payload.Password, user.HashedPassword) {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": fmt.Sprint("invalid password")})
		return
	}

	token, err := auth.GenerateJWT(payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(authCookieName, token, 3600, "", "", false, true)

	c.HTML(http.StatusOK, "index.html", gin.H{"User": user.Email})
}

func Signout(c *gin.Context) {
	if email := GetUserEmail(c); email != "" {
		c.SetCookie(authCookieName, "", -1, "", "", false, true)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func CookieAuthMiddleware(c *gin.Context) {
	token, err := c.Cookie(authCookieName)
	if err != nil {
		c.Next()
		return
	}

	email, isValid := auth.ValidateJWT(token)
	if !isValid {
		c.Next()
		return
	}

	log.Printf("user %s logged in\n", email)
	c.Set(userContextKey, email)
	c.Next()
}

func GetUserEmail(c *gin.Context) string {
	return c.GetString(userContextKey)
}
