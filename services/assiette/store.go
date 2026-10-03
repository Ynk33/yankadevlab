package main

import (
	"bytes"
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
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT data FROM assiette.team_state WHERE team_id = $1`, teamIDFrom(r.Context())).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, []byte("null"))
		return
	}
	if err != nil {
		h.internalError(w, "failed to read state", err)
		return
	}
	writeJSON(w, data)
}

// PatchState applies a JSON merge patch so concurrent edits by team members on different keys don't overwrite each other.
func (h *StoreHandler) PatchState(w http.ResponseWriter, r *http.Request) {
	var patch map[string]any
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.UseNumber()
	if err := dec.Decode(&patch); err != nil || patch == nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	teamID := teamIDFrom(ctx)

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		h.internalError(w, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO assiette.team_state (team_id, data) VALUES ($1, '{}') ON CONFLICT (team_id) DO NOTHING`,
		teamID); err != nil {
		h.internalError(w, "failed to init state", err)
		return
	}

	var raw []byte
	if err := tx.QueryRowContext(ctx,
		`SELECT data FROM assiette.team_state WHERE team_id = $1 FOR UPDATE`, teamID).Scan(&raw); err != nil {
		h.internalError(w, "failed to lock state", err)
		return
	}

	var current any
	dec = json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&current); err != nil {
		h.internalError(w, "failed to decode state", err)
		return
	}

	merged, err := json.Marshal(mergePatch(current, patch))
	if err != nil {
		h.internalError(w, "failed to encode state", err)
		return
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE assiette.team_state SET data = $2, updated_at = now() WHERE team_id = $1`,
		teamID, merged); err != nil {
		h.internalError(w, "failed to save state", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.internalError(w, "failed to commit state", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StoreHandler) ListCustom(w http.ResponseWriter, r *http.Request) {
	var data []byte
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT COALESCE(jsonb_agg(data ORDER BY created_at), '[]'::jsonb) FROM assiette.custom_recipes WHERE team_id = $1`,
		teamIDFrom(r.Context())).Scan(&data)
	if err != nil {
		h.internalError(w, "failed to list custom recipes", err)
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

	teamID := teamIDFrom(r.Context())
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		h.internalError(w, "failed to begin transaction", err)
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
			`INSERT INTO assiette.custom_recipes (team_id, id, data) VALUES ($1, $2, $3) ON CONFLICT (team_id, id) DO NOTHING`,
			teamID, rec.ID, []byte(raw))
		if err != nil {
			h.internalError(w, "failed to add custom recipe", err, "id", rec.ID)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		h.internalError(w, "failed to commit custom recipes", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StoreHandler) internalError(w http.ResponseWriter, msg string, err error, args ...any) {
	h.Log.Error(msg, append([]any{"error", err}, args...)...)
	http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
