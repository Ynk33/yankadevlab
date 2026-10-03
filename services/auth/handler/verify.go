package handler

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Ynk33/yankadevlab/services/auth/token"
)

type VerifyHandler struct {
	Log       *slog.Logger
	JWTSecret string
	LoginURL  string
}

func (h *VerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		rawToken, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || rawToken == "" {
			h.Log.Warn("malformed auth header")
			http.Error(w, `{"error":"malformed auth header"}`, http.StatusUnauthorized)
			return
		}

		claims, err := token.ParseAccessToken(rawToken, h.JWTSecret)
		if err != nil {
			h.Log.Warn("invalid access token", "error", err)
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		h.allow(w, claims)
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		claims, err := token.ParseAccessToken(cookie.Value, h.JWTSecret)
		if err == nil {
			h.allow(w, claims)
			return
		}
		h.Log.Warn("invalid session cookie", "error", err)
	}

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		h.Log.Info("no valid credentials, redirecting to login")
		http.Redirect(w, r, h.loginRedirect(r), http.StatusFound)
		return
	}

	h.Log.Warn("missing credentials")
	http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
}

func (h *VerifyHandler) allow(w http.ResponseWriter, claims *token.Claims) {
	w.Header().Set("X-User-Id", claims.Subject)
	w.Header().Set("X-User-Email", claims.Email)
	w.WriteHeader(http.StatusOK)

	h.Log.Info("token verified", "user_id", claims.Subject)
}

func (h *VerifyHandler) loginRedirect(r *http.Request) string {
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "https"
	}
	original := proto + "://" + r.Header.Get("X-Forwarded-Host") + r.Header.Get("X-Forwarded-Uri")
	return h.LoginURL + "?rd=" + url.QueryEscape(original)
}
