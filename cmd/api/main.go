package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/runtimeninja/ragops/internal/httpapi"
)

func main() {
	addr := getenv("RAGOPS_HTTP_ADDR", ":8080")

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("ragops api listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
