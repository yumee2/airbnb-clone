package httpserver

import (
	"log/slog"
	"net/http"

	"airbnb.com/services/apartment/internal/adapters/http_server/middleware"
	"airbnb.com/services/apartment/internal/domain/entity"
	"airbnb.com/services/apartment/internal/domain/service"

	"github.com/gin-gonic/gin"
)

type ApartmentController interface {
	CreateApartment(ctx *gin.Context)
	ServeImages(ctx *gin.Context)
	GetApartment(ctx *gin.Context)
	DeleteApartment(ctx *gin.Context)
	UpdateApartment(ctx *gin.Context)
}

type apartmentController struct {
	apartmentService service.ApartmentService
	log              *slog.Logger
}

func NewProfileController(logger *slog.Logger, apartmentService service.ApartmentService) ApartmentController {
	return &apartmentController{log: logger, apartmentService: apartmentService}
}

func (c *apartmentController) CreateApartment(ctx *gin.Context) {
	const fn = "adapters.controller.CreateProfile"
	log := c.log.With(
		slog.String("fn", fn),
	)

	var req entity.CreateApartmentRequest

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, _ := ctx.MultipartForm()
	files := form.File["images"]

	apt, err := c.apartmentService.CreateApartment(&req, userID, files)
	if err != nil {
		log.Error("failed to create an apartment", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, apt)
}

func (c *apartmentController) GetApartment(ctx *gin.Context) {

}

func (c *apartmentController) DeleteApartment(ctx *gin.Context) {

}

func (c *apartmentController) UpdateApartment(ctx *gin.Context) {

}

func (c *apartmentController) ServeImages(ctx *gin.Context) {
	filename := "services/apartment/uploads/" + ctx.Param("filename")
	ctx.File(filename)
}
