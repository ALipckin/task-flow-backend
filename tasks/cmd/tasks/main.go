package main

import (
	"log/slog"
	"os"

	"tasks/internal/app"
	"tasks/internal/config"
	"tasks/logger"
)

func main() {
	log, err := logger.NewProduction()
	if err != nil {
		slog.Error("logger init failed", "err", err)
		os.Exit(1)
	}

	cfg := config.LoadFromEnv()
	a := app.New(cfg, log)
	if err := a.Run(); err != nil {
		slog.Error("app error", "err", err)
		os.Exit(1)
	}
}
