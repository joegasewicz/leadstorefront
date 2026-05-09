package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Health struct{}

func (health *Health) Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "platform_web",
		"status":  "ok",
	})
}

func (health *Health) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (health *Health) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (health *Health) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
