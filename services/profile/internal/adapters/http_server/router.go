package httpserver

import (
	"airbnb.com/services/profile/internal/adapters/http_server/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.Engine, profileController ProfileController) {
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.POST("/profiles", profileController.CreateProfile)
	}

	r.GET("/uploads/:filename", profileController.ServeImages)
}
