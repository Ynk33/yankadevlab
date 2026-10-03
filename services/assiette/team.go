package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	inviteTTL         = 7 * 24 * time.Hour
	minPasswordLength = 8
	maxPasswordBytes  = 72
)

var errInvalidInvite = errors.New("invalid or expired invite")

type member struct {
	Email string `json:"email"`
}

type signupRequest struct {
	Token string `json:"token"`
	credentials
}

func (h *StoreHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.QueryContext(r.Context(),
		`SELECT u.email FROM assiette.team_members m JOIN assiette.users u ON u.id = m.user_id
		 WHERE m.team_id = $1 ORDER BY m.joined_at`, teamIDFrom(r.Context()))
	if err != nil {
		h.internalError(w, "failed to list team members", err)
		return
	}
	defer rows.Close()

	members := []member{}
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.Email); err != nil {
			h.internalError(w, "failed to scan team member", err)
			return
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		h.internalError(w, "failed to list team members", err)
		return
	}

	data, _ := json.Marshal(map[string]any{"members": members})
	writeJSON(w, data)
}

func (h *StoreHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	token, err := newToken()
	if err != nil {
		h.internalError(w, "failed to generate invite token", err)
		return
	}

	ctx := r.Context()
	if _, err := h.DB.ExecContext(ctx, `DELETE FROM assiette.invites WHERE expires_at < now()`); err != nil {
		h.internalError(w, "failed to purge expired invites", err)
		return
	}
	if _, err := h.DB.ExecContext(ctx,
		`INSERT INTO assiette.invites (token_hash, team_id, created_by, expires_at) VALUES ($1, $2, $3, $4)`,
		hashToken(token), teamIDFrom(ctx), userIDFrom(ctx), time.Now().Add(inviteTTL)); err != nil {
		h.internalError(w, "failed to store invite", err)
		return
	}

	data, _ := json.Marshal(map[string]string{"token": token})
	writeJSON(w, data)
}

// JoinTeam moves an existing user into the invite's team; their previous team is deleted if left empty.
func (h *StoreHandler) JoinTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, oldTeamID := userIDFrom(ctx), teamIDFrom(ctx)

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		h.internalError(w, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback()

	teamID, err := consumeInvite(ctx, tx, chi.URLParam(r, "token"))
	if errors.Is(err, errInvalidInvite) {
		http.Error(w, `{"error":"invalid or expired invite"}`, http.StatusGone)
		return
	}
	if err != nil {
		h.internalError(w, "failed to consume invite", err)
		return
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE assiette.team_members SET team_id = $2, joined_at = now() WHERE user_id = $1`,
		userID, teamID); err != nil {
		h.internalError(w, "failed to join team", err)
		return
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM assiette.teams t WHERE t.id = $1 AND t.id <> $2
		 AND NOT EXISTS (SELECT 1 FROM assiette.team_members m WHERE m.team_id = t.id)`,
		oldTeamID, teamID); err != nil {
		h.internalError(w, "failed to delete empty team", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.internalError(w, "failed to commit join", err)
		return
	}
	h.Log.Info("user joined team", "user_id", userID, "team_id", teamID)
	w.WriteHeader(http.StatusNoContent)
}

// Signup creates an account from an invite, adds it to the invite's team and signs it in.
func (h *StoreHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	req.Email = normalizeEmail(req.Email)
	if req.Token == "" || !strings.Contains(req.Email, "@") {
		http.Error(w, `{"error":"invite token and valid email are required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < minPasswordLength || len(req.Password) > maxPasswordBytes {
		http.Error(w, `{"error":"password must be between 8 characters and 72 bytes"}`, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.internalError(w, "failed to hash password", err)
		return
	}

	ctx := r.Context()
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		h.internalError(w, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback()

	teamID, err := consumeInvite(ctx, tx, req.Token)
	if errors.Is(err, errInvalidInvite) {
		http.Error(w, `{"error":"invalid or expired invite"}`, http.StatusGone)
		return
	}
	if err != nil {
		h.internalError(w, "failed to consume invite", err)
		return
	}

	var userID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO assiette.users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		req.Email, string(hash)).Scan(&userID)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		http.Error(w, `{"error":"an account already exists for this email"}`, http.StatusConflict)
		return
	}
	if err != nil {
		h.internalError(w, "failed to create user", err)
		return
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO assiette.team_members (user_id, team_id) VALUES ($1, $2)`, userID, teamID); err != nil {
		h.internalError(w, "failed to add team member", err)
		return
	}

	token, expiresAt, err := createSession(ctx, tx, userID)
	if err != nil {
		h.internalError(w, "failed to create session", err, "user_id", userID)
		return
	}

	if err := tx.Commit(); err != nil {
		h.internalError(w, "failed to commit signup", err)
		return
	}
	setSessionCookie(w, token, expiresAt)
	h.Log.Info("user signed up", "user_id", userID, "team_id", teamID)
	w.WriteHeader(http.StatusNoContent)
}

// consumeInvite locks a valid invite, marks it used and returns its team.
func consumeInvite(ctx context.Context, tx *sql.Tx, token string) (string, error) {
	var teamID string
	err := tx.QueryRowContext(ctx,
		`UPDATE assiette.invites SET used_at = now()
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING team_id`, hashToken(token)).Scan(&teamID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errInvalidInvite
	}
	return teamID, err
}
