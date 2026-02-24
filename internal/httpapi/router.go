package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/runtimeninja/ragops/internal/httpapi/handlers"
	"github.com/runtimeninja/ragops/internal/httpapi/middleware"
	"github.com/runtimeninja/ragops/internal/observability"
)

type Deps struct {
	DBPinger func(ctx context.Context) error
	Logger   *slog.Logger
	Metrics  *observability.Metrics

	Docs *handlers.DocumentsHandler
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	if d.Metrics != nil {
		r.Use(middleware.Metrics(d.Metrics))
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if d.DBPinger == nil {
			WriteError(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := d.DBPinger(ctx); err != nil {
			if d.Logger != nil {
				d.Logger.Warn("readyz db ping failed",
					"request_id", middleware.GetRequestID(r.Context()),
					"error", err,
				)
			}
			WriteError(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Post("/documents", d.Docs.Create)
		r.Get("/documents/{id}/status", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			d.Docs.Status(w, r, id)
		})
	})

	return r
}
