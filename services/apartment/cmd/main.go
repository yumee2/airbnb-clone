package main

import (
	"log"
	"log/slog"
	"os"

	httpserver "airbnb.com/services/apartment/internal/adapters/http_server"
	"airbnb.com/services/apartment/internal/adapters/repository"
	"airbnb.com/services/apartment/internal/config"
	"airbnb.com/services/apartment/internal/domain/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	if err := godotenv.Load("services/apartment/.env"); err != nil {
		log.Println("No .env file found")
	}
	cfg := config.MustLoad()

	log := createLogger(cfg.Env)
	log.Info("apartment app just started")

	aptRepo, err := repository.New(cfg)
	if err != nil {
		log.Error("failed to setup database connection")
		os.Exit(1)
	}

	aptService := service.NewApartmentService(aptRepo, "services/apartment/uploads", log)
	r := setUpHttpServer(log, aptService)
	if err := r.Run(cfg.Address); err != nil {
		log.Error("Failed to start server:", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
}

func setUpHttpServer(log *slog.Logger, aptService service.ApartmentService) *gin.Engine {
	r := gin.Default()
	aptController := httpserver.NewProfileController(log, aptService)
	httpserver.SetupProfileRoutes(r, aptController)
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
