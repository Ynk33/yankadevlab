package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

type historyRange struct {
	span time.Duration
	step time.Duration
}

var historyRanges = map[string]historyRange{
	"1h":  {span: time.Hour, step: 15 * time.Second},
	"24h": {span: 24 * time.Hour, step: 5 * time.Minute},
	"7d":  {span: 7 * 24 * time.Hour, step: 30 * time.Minute},
}

type HistoryHandler struct {
	Name   string
	Series map[string]string
	Prom   *prom.Client
	Log    *slog.Logger
}

type historyPoint struct {
	value     float64
	timestamp int64
}

func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rng, ok := historyRanges[r.URL.Query().Get("range")]
	if !ok {
		http.Error(w, `{"error":"invalid range"}`, http.StatusBadRequest)
		return
	}

	end := time.Now().Truncate(rng.step)
	start := end.Add(-rng.span)

	rows := map[int64]map[string]float64{}
	for key, query := range h.Series {
		body, err := h.Prom.QueryRange(r.Context(), query, start, end, rng.step)
		if err != nil {
			h.Log.Error("range query failed", "metric", h.Name, "series", key, "error", err)
			http.Error(w, `{"error":"failed to query metrics"}`, http.StatusBadGateway)
			return
		}

		points, err := parseMatrix(body)
		if err != nil {
			h.Log.Error("range parse failed", "metric", h.Name, "series", key, "error", err)
			http.Error(w, `{"error":"failed to parse metrics"}`, http.StatusBadGateway)
			return
		}

		for _, p := range points {
			row, ok := rows[p.timestamp]
			if !ok {
				row = map[string]float64{"timestamp": float64(p.timestamp)}
				rows[p.timestamp] = row
			}
			row[key] = p.value
		}
	}

	timestamps := make([]int64, 0, len(rows))
	for ts := range rows {
		timestamps = append(timestamps, ts)
	}
	slices.Sort(timestamps)

	res := make([]map[string]float64, 0, len(timestamps))
	for _, ts := range timestamps {
		res = append(res, rows[ts])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func parseMatrix(body []byte) ([]historyPoint, error) {
	var res struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Values [][2]json.RawMessage `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("decode prometheus response: %w", err)
	}
	if res.Status != "success" {
		return nil, fmt.Errorf("prometheus status %q", res.Status)
	}
	if len(res.Data.Result) == 0 {
		return []historyPoint{}, nil
	}

	values := res.Data.Result[0].Values
	points := make([]historyPoint, 0, len(values))
	for _, sample := range values {
		value, ts, err := parseSample(sample)
		if err != nil {
			return nil, err
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		points = append(points, historyPoint{value: value, timestamp: ts})
	}

	return points, nil
}
