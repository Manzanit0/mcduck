package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/users"
)

type UserPayload struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

type RegistrationController struct {
	DB *sqlx.DB
}

func (_ *RegistrationController) GetRegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func (r *RegistrationController) RegisterUser(c *gin.Context) {
	payload := UserPayload{}
	err := c.ShouldBind(&payload)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	_, err = users.Create(c.Request.Context(), r.DB, users.User{Email: payload.Email, Password: payload.Password})
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	err = setCookieAuth(c, payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{"User": payload.Email})
}

func (_ *RegistrationController) GetLoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (r *RegistrationController) LoginUser(c *gin.Context) {
	payload := UserPayload{}
	err := c.ShouldBind(&payload)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	user, err := users.Find(c.Request.Context(), r.DB, payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "invalid email or password"})
		return
	}

	if !auth.CheckPasswordHash(payload.Password, user.HashedPassword) {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "invalid email or password"})
		return
	}

	err = setCookieAuth(c, payload.Email)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{"User": user.Email})
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
