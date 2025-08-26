package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type AuthController interface {
	CreateProfile(ctx *gin.Context)
	GetProfile(ctx *gin.Context)
	DeleteProfile(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	GetYourProfile(ctx *gin.Context)
}

type authController struct {
	//authService service.AuthService
	log *slog.Logger
}

func NewAuthController(logger *slog.Logger) *authController {
	return &authController{log: logger}
}
