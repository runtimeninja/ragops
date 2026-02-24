package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runtimeninja/ragops/internal/app"
	"github.com/runtimeninja/ragops/internal/httpapi"
	"github.com/runtimeninja/ragops/internal/observability"
	"github.com/runtimeninja/ragops/internal/storage"
)

func main() {
	cfg := app.Load()
	ctx := context.Background()

	logger := observability.NewLogger(cfg.Env)
	metrics := observability.NewMetrics()

	db, err := storage.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.NewRouter(httpapi.Deps{
			DBPinger: db.Ping,
			Logger:   logger,
			Metrics:  metrics,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	logger.Info("shutdown complete")
}
