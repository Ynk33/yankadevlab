package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type LoginHandler struct {
	DB                   *sql.DB
	Log                  *slog.Logger
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Decode request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.Warn("invalid request body", "error", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 2. Validate required fields
	if req.Email == "" || req.Password == "" {
		h.Log.Warn("missing email or password")
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	// 3. Look up user by email
	var userID, email, passwordHash string
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT id, email, password_hash FROM users WHERE email = $1`, req.Email,
	).Scan(&userID, &email, &passwordHash)
	if err == sql.ErrNoRows {
		// Timing side-channel mitigation: still compare a dummy hash
		bcrypt.CompareHashAndPassword([]byte("$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), []byte(req.Password))
		h.Log.Info("login failed: unknown email", "email", req.Email)
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.Log.Error("db query failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 4. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		h.Log.Info("login failed: wrong password", "user_id", userID)
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// 5. Generate access token (JWT)
	accessToken, err := token.GenerateAccessToken(userID, email, h.JWTSecret, h.AccessTokenDuration)
	if err != nil {
		h.Log.Error("failed to generate access token", "error", err, "user_id", userID)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// 6. Generate refresh token (opaque) and store hash in DB
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

	// 7. Update last_login_at
	if _, err := h.DB.ExecContext(r.Context(),
		`UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1`, userID,
	); err != nil {
		h.Log.Warn("failed to update last_login_at", "error", err, "user_id", userID)
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

	h.Log.Info("login successful", "user_id", userID)
}
