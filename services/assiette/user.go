package main

import (
	"encoding/json"
	"net/http"
)

type userSettings struct {
	Lang string `json:"lang"`
}

func (h *StoreHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	var s userSettings
	if err := h.DB.QueryRowContext(r.Context(),
		`SELECT lang FROM assiette.users WHERE id = $1`, userIDFrom(r.Context())).Scan(&s.Lang); err != nil {
		h.internalError(w, "failed to read user settings", err)
		return
	}
	data, _ := json.Marshal(s)
	writeJSON(w, data)
}

func (h *StoreHandler) PutMe(w http.ResponseWriter, r *http.Request) {
	var s userSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&s); err != nil || (s.Lang != "fr" && s.Lang != "en") {
		http.Error(w, `{"error":"lang must be fr or en"}`, http.StatusBadRequest)
		return
	}
	if _, err := h.DB.ExecContext(r.Context(),
		`UPDATE assiette.users SET lang = $2, updated_at = now() WHERE id = $1`,
		userIDFrom(r.Context()), s.Lang); err != nil {
		h.internalError(w, "failed to save user settings", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
