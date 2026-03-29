package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ynk33/yankadevlab/services/auth/token"
)

type VerifyHandler struct {
	Log       *slog.Logger
	JWTSecret string
}

func (h *VerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		h.Log.Warn("missing auth header")
		http.Error(w, `{"error":"missing auth header"}`, http.StatusUnauthorized)
		return
	}

	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	if rawToken == "" {
		h.Log.Warn("missing auth token")
		http.Error(w, `{"error":"missing auth token"}`, http.StatusUnauthorized)
		return
	}

	claims, err := token.ParseAccessToken(rawToken, h.JWTSecret)
	if err != nil {
		h.Log.Warn("invalid access token", "error", err)
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("X-User-Id", claims.Subject)
	w.Header().Set("X-User-Email", claims.Email)
	w.WriteHeader(http.StatusOK)

	h.Log.Info("token verified", "user_id", claims.Subject)
}
