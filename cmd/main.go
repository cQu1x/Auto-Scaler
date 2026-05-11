package main

import (
	"log"
	"net/http"

	"github.com/cQu1x/Auto-Scaler/internal/handlers"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	registry := prometheus.NewRegistry()
	metrics := handlers.NewMetrics(registry)
	loader := handlers.NewLoaderHandler(handlers.CPULoader{})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/load", loader.Load)
	mux.Handle("/metrics", promHandler(metrics))

	server := &http.Server{
		Addr:    ":8080",
		Handler: metrics.Middleware(mux),
	}

	log.Println("listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func promHandler(metrics *handlers.Metrics) http.Handler {
	return metrics.Handler()
}
