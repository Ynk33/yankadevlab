package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

const cpuUsageQuery = `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`

type CPUHandler struct {
	Prom *prom.Client
	Log  *slog.Logger
}

type cpuResponse struct {
	UsagePercent float64 `json:"usage_percent"`
	Timestamp    int64   `json:"timestamp"`
}

func (h *CPUHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := h.Prom.Query(r.Context(), cpuUsageQuery)
	if err != nil {
		h.Log.Error("cpu query failed", "error", err)
		http.Error(w, `{"error":"failed to query metrics"}`, http.StatusBadGateway)
		return
	}

	usage, ts, err := parseScalar(body)
	if err != nil {
		h.Log.Error("cpu parse failed", "error", err)
		http.Error(w, `{"error":"failed to parse metrics"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cpuResponse{
		UsagePercent: usage,
		Timestamp:    ts,
	})
}

func parseScalar(body []byte) (float64, int64, error) {
	var res struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value [2]json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, 0, fmt.Errorf("decode prometheus response: %w", err)
	}
	if res.Status != "success" {
		return 0, 0, fmt.Errorf("prometheus status %q", res.Status)
	}
	if len(res.Data.Result) == 0 {
		return 0, 0, fmt.Errorf("empty result")
	}

	var ts float64
	if err := json.Unmarshal(res.Data.Result[0].Value[0], &ts); err != nil {
		return 0, 0, fmt.Errorf("decode timestamp: %w", err)
	}

	var valueStr string
	if err := json.Unmarshal(res.Data.Result[0].Value[1], &valueStr); err != nil {
		return 0, 0, fmt.Errorf("decode value: %w", err)
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse value %q: %w", valueStr, err)
	}

	return value, int64(ts), nil
}
