package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/auth"
)

func LandingPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"User": auth.GetUserEmail(c),
	})
}
