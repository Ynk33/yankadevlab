package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

type ScalarHandler struct {
	Name  string
	Query string
	Prom  *prom.Client
	Log   *slog.Logger
}

type scalarResponse struct {
	UsagePercent float64 `json:"usage_percent"`
	Timestamp    int64   `json:"timestamp"`
}

func (h *ScalarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := h.Prom.Query(r.Context(), h.Query)
	if err != nil {
		h.Log.Error("query failed", "metric", h.Name, "error", err)
		http.Error(w, `{"error":"failed to query metrics"}`, http.StatusBadGateway)
		return
	}

	usage, ts, err := parseScalar(body)
	if err != nil {
		h.Log.Error("parse failed", "metric", h.Name, "error", err)
		http.Error(w, `{"error":"failed to parse metrics"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scalarResponse{
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
