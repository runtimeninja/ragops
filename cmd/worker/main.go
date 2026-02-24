package main

import (
	"os"

	"github.com/runtimeninja/ragops/internal/app"
	"github.com/runtimeninja/ragops/internal/jobs"
	"github.com/runtimeninja/ragops/internal/observability"
)

func main() {
	cfg := app.Load()

	logger := observability.NewLogger(cfg.Env)

	w := jobs.NewWorker(logger)
	if err := w.Run(cfg.RedisAddr); err != nil {
		logger.Error("worker failed", "error", err)
		os.Exit(1)
	}
}
