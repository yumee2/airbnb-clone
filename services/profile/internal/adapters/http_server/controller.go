package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"airbnb.com/services/profile/internal/adapters/http_server/middleware"
	"airbnb.com/services/profile/internal/domain/entity"
	"airbnb.com/services/profile/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type ProfileController interface {
	CreateProfile(ctx *gin.Context)
	ServeImages(ctx *gin.Context)
	// GetProfile(ctx *gin.Context)
	// DeleteProfile(ctx *gin.Context)
	// UpdateProfile(ctx *gin.Context)
	// GetYourProfile(ctx *gin.Context)
}

type profileController struct {
	profileService service.ProfileService
	log            *slog.Logger
}

func NewProfileController(logger *slog.Logger, profileService service.ProfileService) ProfileController {
	return &profileController{log: logger, profileService: profileService}
}

func (c *profileController) CreateProfile(ctx *gin.Context) {
	const fn = "adapters.controller.CreateProfile"
	log := c.log.With(
		slog.String("fn", fn),
	)

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		log.Error("failed to parse form data", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	request := &entity.CreateProfileRequest{
		PhoneNumber: ctx.Request.FormValue("phone_number"),
		Name:        ctx.Request.FormValue("name"),
		Surname:     ctx.Request.FormValue("surname"),
	}

	dobStr := ctx.Request.FormValue("date_of_birth")
	if dobStr != "" {
		dob, err := time.Parse("02-01-2006", dobStr)
		if err != nil {
			log.Error("failed to parse date of birth", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use DD-MM-YYYY"})
			return
		}
		request.DateOfBirth = dob
	}

	imageFile, err := ctx.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
		return
	}

	if request.PhoneNumber == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}
	if request.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	if request.DateOfBirth.IsZero() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Date of birth is required"})
		return
	}

	profile, err := c.profileService.CreateProfile(request, userID, imageFile)
	if err != nil {
		log.Error("failed to create a profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, profile)
}

func (c *profileController) ServeImages(ctx *gin.Context) {
	filename := "services/profile/uploads/" + ctx.Param("filename")
	ctx.File(filename)
}
