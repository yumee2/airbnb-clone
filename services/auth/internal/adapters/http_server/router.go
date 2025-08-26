package http_server

import (
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, authController AuthController) {
	urlGroup := r.Group("/auth")
	{
		urlGroup.POST("/register", authController.Register)
		urlGroup.POST("/login", authController.Login)
		urlGroup.POST("/refresh", authController.Refresh)
	}
}
