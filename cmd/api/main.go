package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runtimeninja/ragops/internal/app"
	"github.com/runtimeninja/ragops/internal/httpapi"
	"github.com/runtimeninja/ragops/internal/storage"
)

func main() {
	cfg := app.Load()
	ctx := context.Background()

	db, err := storage.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer db.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(httpapi.Deps{DBPinger: db.Ping}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("ragops api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	log.Printf("shutdown complete")
}
