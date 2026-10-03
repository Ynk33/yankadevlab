package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
)

type LogoutHandler struct {
	DB           *sql.DB
	Log          *slog.Logger
	CookieDomain string
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Clear the session cookie, even without a refresh token
	clearSessionCookie(w, h.CookieDomain)

	// 2. Read the refresh_token cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.Log.Info("missing refresh token", "error", err)
		http.Error(w, `{"error":"missing refresh token"}`, http.StatusUnauthorized)
		return
	}

	// 3. Hash the token
	hash := token.HashToken(cookie.Value)

	// 4. Revoke the token
	if _, err := h.DB.ExecContext(r.Context(),
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`, hash,
	); err != nil {
		h.Log.Error("db query failed", "error", err)
	}

	// 5. Delete the refresh_token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Path:     "/", // scoped to /refresh later if needed
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	h.Log.Info("logout successful")
	w.WriteHeader(http.StatusNoContent)
}
