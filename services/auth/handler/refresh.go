package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
)

type RefreshHandler struct {
	DB                   *sql.DB
	Log                  *slog.Logger
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func (h *RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Read the refresh_token cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.Log.Info("missing refresh token", "error", err)
		http.Error(w, `{"error":"missing refresh token"}`, http.StatusUnauthorized)
		return
	}

	// 2. Hash the token
	hash := token.HashToken(cookie.Value)

	// 3. Fetch matching userID
	var userID string
	err = h.DB.QueryRowContext(r.Context(),
		`SELECT user_id FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`, hash,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		h.Log.Info("refresh failed: invalid or expired refresh token", "error", err)
		http.Error(w, `{"error":"invalid or expired refresh token"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.Log.Error("db query failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 4. Revoke old token
	if _, err := h.DB.ExecContext(r.Context(),
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`, hash,
	); err != nil {
		h.Log.Error("db query failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 5. Fetch user's email
	var email string
	err = h.DB.QueryRowContext(r.Context(),
		`SELECT email FROM users WHERE id = $1`, userID,
	).Scan(&email)
	if err == sql.ErrNoRows {
		h.Log.Error("refresh failed: user not found", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if err != nil {
		h.Log.Error("db query failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 6. Generate access token (JWT)
	accessToken, err := token.GenerateAccessToken(userID, email, h.JWTSecret, h.AccessTokenDuration)
	if err != nil {
		h.Log.Error("failed to generate access token", "error", err, "user_id", userID)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 7. Generate refresh token (opaque) and store hash in DB
	rawRefresh, refreshHash, err := token.GenerateRefreshToken()
	if err != nil {
		h.Log.Error("failed to generate refresh token", "error", err, "user_id", userID)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(h.RefreshTokenDuration)
	_, err = h.DB.ExecContext(r.Context(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, refreshHash, expiresAt,
	)
	if err != nil {
		h.Log.Error("failed to store refresh token", "error", err, "user_id", userID)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 8. Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rawRefresh,
		Path:     "/", // scoped to /refresh later if needed
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expiresAt,
	})

	// 9. Return access token in response body
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{AccessToken: accessToken})

	h.Log.Info("refresh token successful", "user_id", userID)
}
