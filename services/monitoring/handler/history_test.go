package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Ynk33/yankadevlab/services/monitoring/prom"
)

func TestParseMatrix(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    []historyPoint
		wantErr bool
	}{
		{
			name: "success",
			body: `{"status":"success","data":{"result":[{"values":[[1700000000,"12.5"],[1700000015,"13"]]}]}}`,
			want: []historyPoint{{value: 12.5, timestamp: 1700000000}, {value: 13, timestamp: 1700000015}},
		},
		{
			name: "skips NaN",
			body: `{"status":"success","data":{"result":[{"values":[[1700000000,"NaN"],[1700000015,"1"]]}]}}`,
			want: []historyPoint{{value: 1, timestamp: 1700000015}},
		},
		{name: "error status", body: `{"status":"error"}`, wantErr: true},
		{name: "empty result", body: `{"status":"success","data":{"result":[]}}`, want: []historyPoint{}},
		{
			name:    "bad value",
			body:    `{"status":"success","data":{"result":[{"values":[[1700000000,"abc"]]}]}}`,
			wantErr: true,
		},
		{name: "invalid json", body: `{`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMatrix([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHistoryHandler(t *testing.T) {
	responses := map[string]string{
		"rx_query": `{"status":"success","data":{"result":[{"values":[[100,"1"],[115,"2"]]}]}}`,
		"tx_query": `{"status":"success","data":{"result":[{"values":[[115,"3"],[130,"4"]]}]}}`,
	}
	promServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responses[r.URL.Query().Get("query")]))
	}))
	defer promServer.Close()

	h := &HistoryHandler{
		Name:   "network",
		Series: map[string]string{"rx": "rx_query", "tx": "tx_query"},
		Prom:   prom.NewClient(promServer.URL),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   []map[string]float64
	}{
		{name: "missing range", query: "", wantStatus: http.StatusBadRequest},
		{name: "unknown range", query: "?range=1y", wantStatus: http.StatusBadRequest},
		{
			name:       "merges series by timestamp",
			query:      "?range=1h",
			wantStatus: http.StatusOK,
			wantBody: []map[string]float64{
				{"timestamp": 100, "rx": 1},
				{"timestamp": 115, "rx": 2, "tx": 3},
				{"timestamp": 130, "tx": 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics/network/history"+tt.query, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody == nil {
				return
			}
			var got []map[string]float64
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.wantBody) {
				t.Errorf("got %v, want %v", got, tt.wantBody)
			}
		})
	}
}
