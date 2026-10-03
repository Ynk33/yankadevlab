package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "assiette_session"
	sessionTTL        = 30 * 24 * time.Hour
	sessionRenewBelow = 15 * 24 * time.Hour
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ctxKey int

const (
	userIDKey ctxKey = iota
	teamIDKey
)

func userIDFrom(ctx context.Context) string { return ctx.Value(userIDKey).(string) }
func teamIDFrom(ctx context.Context) string { return ctx.Value(teamIDKey).(string) }

// WithSession authenticates the request from the session cookie and stores the user and team in the context.
func (h *StoreHandler) WithSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		hash := hashToken(cookie.Value)
		var userID, teamID string
		var expiresAt time.Time
		err = h.DB.QueryRowContext(ctx,
			`SELECT s.user_id, m.team_id, s.expires_at FROM assiette.sessions s
			 JOIN assiette.team_members m ON m.user_id = s.user_id
			 WHERE s.token_hash = $1 AND s.expires_at > now()`, hash).Scan(&userID, &teamID, &expiresAt)
		if errors.Is(err, sql.ErrNoRows) {
			clearSessionCookie(w)
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if err != nil {
			h.internalError(w, "failed to read session", err)
			return
		}

		if time.Until(expiresAt) < sessionRenewBelow {
			newExpiry := time.Now().Add(sessionTTL)
			if _, err := h.DB.ExecContext(ctx,
				`UPDATE assiette.sessions SET expires_at = $2 WHERE token_hash = $1`, hash, newExpiry); err != nil {
				h.Log.Warn("failed to renew session", "error", err, "user_id", userID)
			} else {
				setSessionCookie(w, cookie.Value, newExpiry)
			}
		}

		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = context.WithValue(ctx, teamIDKey, teamID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *StoreHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req credentials
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	email := normalizeEmail(req.Email)

	ctx := r.Context()
	var userID, passwordHash string
	err := h.DB.QueryRowContext(ctx,
		`SELECT id, password_hash FROM assiette.users WHERE email = $1`, email).Scan(&userID, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		bcrypt.CompareHashAndPassword([]byte("$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), []byte(req.Password))
		h.Log.Info("login failed: unknown email")
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.internalError(w, "failed to look up user", err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		h.Log.Info("login failed: wrong password", "user_id", userID)
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, expiresAt, err := createSession(ctx, h.DB, userID)
	if err != nil {
		h.internalError(w, "failed to create session", err, "user_id", userID)
		return
	}
	setSessionCookie(w, token, expiresAt)
	h.Log.Info("login successful", "user_id", userID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *StoreHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if _, err := h.DB.ExecContext(r.Context(),
			`DELETE FROM assiette.sessions WHERE token_hash = $1`, hashToken(cookie.Value)); err != nil {
			h.Log.Warn("failed to delete session", "error", err)
		}
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// createSession stores a new session for userID, purging expired ones, and returns the raw token.
func createSession(ctx context.Context, db execer, userID string) (string, time.Time, error) {
	token, err := newToken()
	if err != nil {
		return "", time.Time{}, err
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM assiette.sessions WHERE expires_at < now()`); err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(sessionTTL)
	if _, err := db.ExecContext(ctx,
		`INSERT INTO assiette.sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hashToken(token), userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
