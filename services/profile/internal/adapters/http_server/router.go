package httpserver

import (
	"airbnb-clone/profile/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.Engine, profileController ProfileController) {
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.POST("/profile", profileController.CreateProfile)
		authGroup.GET("/user/me", profileController.GetYourProfile)
		authGroup.DELETE("/profile", profileController.DeleteProfile)
		authGroup.PUT("/profile", profileController.UpdateProfile)
	}

	r.GET("/user/:id", profileController.GetProfile)
	r.GET("/uploads/:filename", profileController.ServeImages)
}
