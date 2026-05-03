package routes

import "github.com/gin-gonic/gin"

func Register(app *gin.Engine) {
	registerHealth(app)
}
