package main

import (
	"airbnb-clone/auth/internal/adapters/http_server"
	"airbnb-clone/auth/internal/adapters/repository"
	"airbnb-clone/auth/internal/config"
	"airbnb-clone/auth/internal/domain/service"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	if err := godotenv.Load("services/auth/.env"); err != nil {
		log.Println("No .env file found")
	}
	cfg := config.MustLoad()

	log := createLogger(cfg.Env)
	log.Info("auth app just started")

	authRepo, err := repository.New(cfg)
	if err != nil {
		log.Error("failed to setup database connection")
		os.Exit(1)
	}

	authService := service.NewAuthService(authRepo, log)
	r := setUpHttpServer(log, authService)
	if err := r.Run(cfg.Address); err != nil {
		log.Error("Failed to start server:", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
}

func setUpHttpServer(log *slog.Logger, authService service.AuthService) *gin.Engine {
	r := gin.Default()
	authController := http_server.NewAuthController(log, authService)
	http_server.SetupAuthRoutes(r, authController)
	return r
}

func createLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
