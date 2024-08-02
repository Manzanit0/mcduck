package auth

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/authn"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	signingKey = "AllYourBase"
	threeDays  = time.Hour * 72
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(email string) (string, error) {
	mySigningKey := []byte(signingKey)

	now := time.Now()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(threeDays)),
		Issuer:    "mcduck",
		Subject:   email,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

func ValidateJWT(tokenString string) (string, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil {
		return "", false
	}

	//  && claims.NotBefore.After(time.Now()) && claims.ExpiresAt.After(time.Now())
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, true
	}

	return "", false
}

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

// ConnectMiddleware enforces authentication on all procedures except those
// provided in the allow list.
// Copied from https://github.com/connectrpc/authn-go/pull/12
func ConnectMiddleware(allowed ...string) *authn.Middleware {
	allowList := map[string]struct{}{}
	for _, s := range allowed {
		allowList[s] = struct{}{}
	}

	authenticate := func(_ context.Context, req authn.Request) (any, error) {
		token, ok := bearerToken(req)
		if !ok {
			// We'll allow unauthenticated access to the ping procedure.
			if _, ok := allowList[req.Procedure()]; ok {
				return nil, nil // no authentication required
			}

			err := authn.Errorf("invalid authorization")
			err.Meta().Set("WWW-Authenticate", "Bearer")
			return nil, err
		}

		user, isValid := ValidateJWT(token)
		if !isValid {
			return nil, authn.Errorf("invalid token")
		}

		return user, nil
	}

	return authn.NewMiddleware(authenticate)
}

func bearerToken(r authn.Request) (string, bool) {
	auth := r.Header().Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", false
	}
	return auth[len(prefix):], true
}
