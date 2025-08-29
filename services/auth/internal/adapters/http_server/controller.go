package http_server

import (
	"errors"
	"log/slog"
	"net/http"

	"airbnb.com/services/auth/internal/domain/service"

	"github.com/gin-gonic/gin"
)

type AuthController interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	Refresh(ctx *gin.Context)
}

type authController struct {
	authService service.AuthService
	log         *slog.Logger
}

func NewAuthController(logger *slog.Logger, authServ service.AuthService) *authController {
	return &authController{log: logger, authService: authServ}
}

type JWTTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *authController) Register(ctx *gin.Context) {
	const fn = "adapters.controller.Register"
	log := c.log.With(
		slog.String("fn", fn),
	)

	var request registerRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtTokenPair, err := c.authService.RegisterNewUser(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExist) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User with provided email already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, JWTTokenResponse{
		AccessToken:  jwtTokenPair.AccessToken,
		RefreshToken: jwtTokenPair.RefreshToken,
	})
}

func (c *authController) Login(ctx *gin.Context) {
	const fn = "adapters.controller.Login"
	log := c.log.With(
		slog.String("fn", fn),
	)

	var request loginRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtTokenPair, err := c.authService.LoginExistingUser(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User with provided email was not found"})
			return
		}

		if errors.Is(err, service.ErrInvalidPassword) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, JWTTokenResponse{
		AccessToken:  jwtTokenPair.AccessToken,
		RefreshToken: jwtTokenPair.RefreshToken,
	})
}

func (c *authController) Refresh(ctx *gin.Context) {
	const fn = "adapters.controller.Refresh"
	log := c.log.With(
		slog.String("fn", fn),
	)

	var request refreshRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := c.authService.ValidateRefreshToken(request.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrRefreshTokenNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token not found"})
			return
		}
		if errors.Is(err, service.ErrRefreshTokenExpired) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
