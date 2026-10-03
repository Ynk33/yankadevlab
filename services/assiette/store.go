package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

const maxBodyBytes = 1 << 20

type StoreHandler struct {
	DB  *sql.DB
	Log *slog.Logger
}

type customRecipe struct {
	ID string `json:"id"`
}

func (h *StoreHandler) GetState(w http.ResponseWriter, r *http.Request) {
	var data []byte
	err := h.DB.QueryRowContext(r.Context(), `SELECT data FROM assiette.state WHERE id = 1`).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, []byte("null"))
		return
	}
	if err != nil {
		h.Log.Error("failed to read state", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

func (h *StoreHandler) PutState(w http.ResponseWriter, r *http.Request) {
	var data json.RawMessage
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&data); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	_, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO assiette.state (id, data, updated_at) VALUES (1, $1, now())
		 ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, updated_at = now()`,
		[]byte(data))
	if err != nil {
		h.Log.Error("failed to save state", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StoreHandler) ListCustom(w http.ResponseWriter, r *http.Request) {
	var data []byte
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT COALESCE(jsonb_agg(data ORDER BY created_at), '[]'::jsonb) FROM assiette.custom_recipes`).Scan(&data)
	if err != nil {
		h.Log.Error("failed to list custom recipes", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

func (h *StoreHandler) AddCustom(w http.ResponseWriter, r *http.Request) {
	var recipes []json.RawMessage
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&recipes); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.Log.Error("failed to begin transaction", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, raw := range recipes {
		var rec customRecipe
		if err := json.Unmarshal(raw, &rec); err != nil || rec.ID == "" {
			http.Error(w, `{"error":"recipe id is required"}`, http.StatusBadRequest)
			return
		}
		_, err := tx.ExecContext(r.Context(),
			`INSERT INTO assiette.custom_recipes (id, data) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
			rec.ID, []byte(raw))
		if err != nil {
			h.Log.Error("failed to add custom recipe", "error", err, "id", rec.ID)
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		h.Log.Error("failed to commit custom recipes", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
