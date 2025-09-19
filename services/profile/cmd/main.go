package main

import (
	httpserver "airbnb-clone/profile/internal/adapters/http_server"
	"airbnb-clone/profile/internal/adapters/repository"
	"airbnb-clone/profile/internal/config"
	"airbnb-clone/profile/internal/domain/service"

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
	if err := godotenv.Load("services/profile/.env"); err != nil {
		log.Println("No .env file found")
	}
	cfg := config.MustLoad()

	log := createLogger(cfg.Env)
	log.Info("profile app just started")

	profileRepo, err := repository.New(cfg)
	if err != nil {
		log.Error("failed to setup database connection")
		os.Exit(1)
	}

	profileService := service.NewProfileService(profileRepo, log, "services/profile/uploads")
	r := setUpHttpServer(log, profileService)
	if err := r.Run(cfg.Address); err != nil {
		log.Error("Failed to start server:", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
}

func setUpHttpServer(log *slog.Logger, profileService service.ProfileService) *gin.Engine {
	r := gin.Default()
	profileController := httpserver.NewProfileController(log, profileService)
	httpserver.SetupProfileRoutes(r, profileController)
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
