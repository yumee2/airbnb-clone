package main

import (
	"log"
	"log/slog"
	"os"

	"airbnb.com/services/profile/internal/config"
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
