package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/runtimeninja/ragops/internal/observability"
)

func Metrics(m *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rr := &statusRecorder{ResponseWriter: w, status: 200}

			next.ServeHTTP(rr, r)

			route := routePattern(r)
			m.HTTPRequests.WithLabelValues(route, r.Method, strconv.Itoa(rr.status)).Inc()
			m.HTTPDuration.WithLabelValues(route, r.Method).Observe(time.Since(start).Seconds())
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		p := rc.RoutePattern()
		if p != "" {
			return p
		}
	}
	return "unknown"
}
