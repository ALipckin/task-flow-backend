package main

import (
	"log/slog"
	"os"

	"tasks/internal/app"
	"tasks/logger"
)

func main() {
	logger.Init()

	a := app.New()
	if err := a.Run(); err != nil {
		slog.Error("app error", "err", err)
		os.Exit(1)
	}
}
