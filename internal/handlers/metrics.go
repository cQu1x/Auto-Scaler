package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type Metrics struct {
	requestsTotal      *prometheus.CounterVec
	requestDuration    *prometheus.HistogramVec
	droppedRequestsTot *prometheus.CounterVec
	gatherer           prometheus.Gatherer
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of handled HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		droppedRequestsTot: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_dropped_total",
				Help: "Total number of HTTP requests canceled or timed out before completion.",
			},
			[]string{"method", "path", "reason"},
		),
		gatherer: registry,
	}

	registry.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.droppedRequestsTot,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		path := r.URL.Path
		method := r.Method
		status := strconv.Itoa(rw.statusCode)

		m.requestsTotal.WithLabelValues(method, path, status).Inc()
		m.requestDuration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())

		if err := r.Context().Err(); err != nil {
			m.droppedRequestsTot.WithLabelValues(method, path, err.Error()).Inc()
		}
	})
}
