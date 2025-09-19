package httpserver

import (
	"airbnb-clone/profile/internal/adapters/http_server/middleware"
	"airbnb-clone/profile/internal/domain/entity"
	"airbnb-clone/profile/internal/domain/service"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ProfileController interface {
	CreateProfile(ctx *gin.Context)
	ServeImages(ctx *gin.Context)
	GetProfile(ctx *gin.Context)
	DeleteProfile(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	GetYourProfile(ctx *gin.Context)
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

func (c *profileController) GetProfile(ctx *gin.Context) {
	const fn = "adapters.controller.GetProfile"
	log := c.log.With(
		slog.String("fn", fn),
	)

	userID := ctx.Param("id")
	if userID == "" {
		log.Error("user id was not provided")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID was not provided"})
		return
	}
	profileResponse, err := c.profileService.GetProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User with provided ID was not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, profileResponse)
}

func (c *profileController) GetYourProfile(ctx *gin.Context) {
	const fn = "adapters.controller.GetYourProfile"
	log := c.log.With(
		slog.String("fn", fn),
	)

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error("failed to get user id out of context", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	profileResponse, err := c.profileService.GetYourProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User with provided ID was not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, profileResponse)
}

func (c *profileController) DeleteProfile(ctx *gin.Context) {
	const fn = "adapters.controller.GetYourProfile"
	log := c.log.With(
		slog.String("fn", fn),
	)

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error("failed to get user id out of context", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err = c.profileService.DeleteProfile(userID); err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Profile with provided ID not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (c *profileController) UpdateProfile(ctx *gin.Context) {
	const fn = "adapters.controller.UpdateProfile"
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

	request := &entity.UpdateProfileRequest{}

	if phoneNumber := ctx.Request.FormValue("phone_number"); phoneNumber != "" {
		request.PhoneNumber = &phoneNumber
	}
	if name := ctx.Request.FormValue("name"); name != "" {
		request.Name = &name
	}
	if surname := ctx.Request.FormValue("surname"); surname != "" {
		request.Surname = &surname
	}

	if dobStr := ctx.Request.FormValue("date_of_birth"); dobStr != "" {
		dob, err := time.Parse("02-01-2006", dobStr)
		if err != nil {
			log.Error("failed to parse date of birth", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use DD-MM-YYYY"})
			return
		}
		request.DateOfBirth = &dob
	}

	imageFile, err := ctx.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
		return
	}

	if request.PhoneNumber == nil && request.Name == nil && request.Surname == nil &&
		request.DateOfBirth == nil && imageFile == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	profile, err := c.profileService.UpdateProfile(userID, request, imageFile)
	if err != nil {
		log.Error("failed to update profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

func (c *profileController) ServeImages(ctx *gin.Context) {
	filename := "services/profile/uploads/" + ctx.Param("filename")
	ctx.File(filename)
}
