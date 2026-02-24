package observability

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ragops_http_requests_total",
			Help: "Total HTTP requests",
		}, []string{"route", "method", "status"}),

		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "ragops_http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
		}, []string{"route", "method"}),
	}

	prometheus.MustRegister(m.HTTPRequests, m.HTTPDuration)
	return m
}
