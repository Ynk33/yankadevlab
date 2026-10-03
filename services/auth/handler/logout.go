package handler

import (
	"log/slog"
	"net/http"
)

type LogoutHandler struct {
	Log          *slog.Logger
	CookieDomain string
	LoginURL     string
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w, h.CookieDomain)
	http.Redirect(w, r, h.LoginURL, http.StatusSeeOther)

	h.Log.Info("logout successful")
}
