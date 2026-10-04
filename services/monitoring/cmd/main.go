package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/Ynk33/yankadevlab/services/monitoring/handler"
	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

const (
	cpuQuery   = `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[30s])) * 100)`
	ramQuery   = `100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)`
	diskQuery  = `100 * (1 - node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})`
	netRxQuery = `sum(irate(node_network_receive_bytes_total{device!="lo"}[30s]))`
	netTxQuery = `sum(irate(node_network_transmit_bytes_total{device!="lo"}[30s]))`
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

	scalar := func(name, query string) http.HandlerFunc {
		h := &handler.ScalarHandler{Name: name, Query: query, Prom: promClient, Log: logger}
		return h.ServeHTTP
	}

	history := func(name string, series map[string]string) http.HandlerFunc {
		h := &handler.HistoryHandler{Name: name, Series: series, Prom: promClient, Log: logger}
		return h.ServeHTTP
	}

	networkHandler := &handler.NetworkHandler{
		RxQuery: netRxQuery,
		TxQuery: netTxQuery,
		Prom:    promClient,
		Log:     logger,
	}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/metrics/cpu", scalar("cpu", cpuQuery))
	r.Get("/metrics/ram", scalar("ram", ramQuery))
	r.Get("/metrics/disk", scalar("disk", diskQuery))
	r.Get("/metrics/network", networkHandler.ServeHTTP)
	r.Get("/metrics/cpu/history", history("cpu", map[string]string{"value": cpuQuery}))
	r.Get("/metrics/ram/history", history("ram", map[string]string{"value": ramQuery}))
	r.Get("/metrics/disk/history", history("disk", map[string]string{"value": diskQuery}))
	r.Get("/metrics/network/history", history("network", map[string]string{"rx": netRxQuery, "tx": netTxQuery}))

	logger.Info("monitoring service listening", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
