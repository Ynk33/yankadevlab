package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

type NetworkHandler struct {
	RxQuery string
	TxQuery string
	Prom    *prom.Client
	Log     *slog.Logger
}

type networkResponse struct {
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
	Timestamp     int64   `json:"timestamp"`
}

func (h *NetworkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rxBody, err := h.Prom.Query(r.Context(), h.RxQuery)
	if err != nil {
		h.Log.Error("network rx query failed", "error", err)
		http.Error(w, `{"error":"failed to query metrics"}`, http.StatusBadGateway)
		return
	}
	rx, ts, err := parseScalar(rxBody)
	if err != nil {
		h.Log.Error("network rx parse failed", "error", err)
		http.Error(w, `{"error":"failed to parse metrics"}`, http.StatusBadGateway)
		return
	}

	txBody, err := h.Prom.Query(r.Context(), h.TxQuery)
	if err != nil {
		h.Log.Error("network tx query failed", "error", err)
		http.Error(w, `{"error":"failed to query metrics"}`, http.StatusBadGateway)
		return
	}
	tx, _, err := parseScalar(txBody)
	if err != nil {
		h.Log.Error("network tx parse failed", "error", err)
		http.Error(w, `{"error":"failed to parse metrics"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(networkResponse{
		RxBytesPerSec: rx,
		TxBytesPerSec: tx,
		Timestamp:     ts,
	})
}
