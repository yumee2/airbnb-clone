package httpserver

import (
	"airbnb.com/services/apartment/internal/adapters/http_server/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.Engine, apartmentController ApartmentController) {
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.POST("/apartment", apartmentController.CreateApartment)
	}
	r.PUT("/apartment/:id", apartmentController.UpdateApartment)
	r.GET("/apartment/:id", apartmentController.GetApartment)
	r.DELETE("/apartment/:id", apartmentController.DeleteApartment)
	r.GET("/uploads/:filename", apartmentController.ServeImages)
}
