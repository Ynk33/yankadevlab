package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/Ynk33/yankadevlab/services/monitoring/handler"
	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://prometheus:9090"
	}

	promClient := prom.NewClient(prometheusURL)

	cpuHandler := &handler.CPUHandler{
		Prom: promClient,
		Log:  logger,
	}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/metrics/cpu", cpuHandler.ServeHTTP)

	logger.Info("monitoring service listening", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
