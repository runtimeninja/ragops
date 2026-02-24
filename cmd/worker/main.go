package main

import (
	"context"
	"os"

	"github.com/runtimeninja/ragops/internal/app"
	"github.com/runtimeninja/ragops/internal/ingest"
	"github.com/runtimeninja/ragops/internal/jobs"
	"github.com/runtimeninja/ragops/internal/observability"
	"github.com/runtimeninja/ragops/internal/storage"
)

func main() {
	cfg := app.Load()
	ctx := context.Background()

	logger := observability.NewLogger(cfg.Env)

	db, err := storage.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ing := ingest.NewService(db.Pool)
	processor := jobs.NewProcessor(ing)

	w := jobs.NewWorker(logger, processor)
	if err := w.Run(cfg.RedisAddr); err != nil {
		logger.Error("worker failed", "error", err)
		os.Exit(1)
	}
}
