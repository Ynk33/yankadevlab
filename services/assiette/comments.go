package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

const maxCommentRunes = 2000

// commentJSON builds the comment payload from rows aliased c (recipe_comments) and u (users); $1 must be the current user id.
const commentJSON = `json_build_object('id', c.id, 'author', u.email, 'body', c.body, 'createdAt', c.created_at, 'updatedAt', c.updated_at, 'mine', c.user_id = $1)`

type commentInput struct {
	Body string `json:"body"`
}

// commentBody trims the comment text and reports whether it is non-empty and within maxCommentRunes.
func commentBody(s string) (string, bool) {
	s = strings.TrimSpace(s)
	return s, s != "" && utf8.RuneCountInString(s) <= maxCommentRunes
}

func decodeComment(w http.ResponseWriter, r *http.Request) (string, bool) {
	var in commentInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&in); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return "", false
	}
	body, ok := commentBody(in.Body)
	if !ok {
		http.Error(w, `{"error":"comment must be 1 to 2000 characters"}`, http.StatusBadRequest)
	}
	return body, ok
}

func (h *StoreHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data []byte
	err := h.DB.QueryRowContext(ctx,
		`SELECT COALESCE(json_agg(`+commentJSON+` ORDER BY c.created_at), '[]'::json)
		 FROM assiette.recipe_comments c JOIN assiette.users u ON u.id = c.user_id
		 WHERE c.team_id = $2 AND c.recipe_id = $3`,
		userIDFrom(ctx), teamIDFrom(ctx), chi.URLParam(r, "id")).Scan(&data)
	if err != nil {
		h.internalError(w, "failed to list comments", err)
		return
	}
	writeJSON(w, data)
}

func (h *StoreHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeComment(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	var data []byte
	err := h.DB.QueryRowContext(ctx,
		`WITH c AS (
		   INSERT INTO assiette.recipe_comments (user_id, team_id, recipe_id, body) VALUES ($1, $2, $3, $4) RETURNING *
		 )
		 SELECT `+commentJSON+` FROM c JOIN assiette.users u ON u.id = c.user_id`,
		userIDFrom(ctx), teamIDFrom(ctx), chi.URLParam(r, "id"), body).Scan(&data)
	if err != nil {
		h.internalError(w, "failed to add comment", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// EditComment only updates comments written by the current user; the id is compared as text so a malformed id is a 404, not a cast error.
func (h *StoreHandler) EditComment(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeComment(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	var data []byte
	err := h.DB.QueryRowContext(ctx,
		`WITH c AS (
		   UPDATE assiette.recipe_comments SET body = $4, updated_at = now()
		   WHERE user_id = $1 AND team_id = $2 AND id::text = $3 RETURNING *
		 )
		 SELECT `+commentJSON+` FROM c JOIN assiette.users u ON u.id = c.user_id`,
		userIDFrom(ctx), teamIDFrom(ctx), chi.URLParam(r, "cid"), body).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		h.internalError(w, "failed to edit comment", err)
		return
	}
	writeJSON(w, data)
}

func (h *StoreHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res, err := h.DB.ExecContext(ctx,
		`DELETE FROM assiette.recipe_comments WHERE user_id = $1 AND team_id = $2 AND id::text = $3`,
		userIDFrom(ctx), teamIDFrom(ctx), chi.URLParam(r, "cid"))
	if err != nil {
		h.internalError(w, "failed to delete comment", err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
