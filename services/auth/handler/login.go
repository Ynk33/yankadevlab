package handler

import (
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	DB              *sql.DB
	Log             *slog.Logger
	JWTSecret       string
	SessionDuration time.Duration
	CookieDomain    string
	DefaultRedirect string
	Template        *template.Template
}

type loginPage struct {
	Redirect string
	Email    string
	Error    string
}

func (h *LoginHandler) Show(w http.ResponseWriter, r *http.Request) {
	redirect := safeRedirect(r.URL.Query().Get("rd"), h.CookieDomain, h.DefaultRedirect)

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if _, err := token.ParseAccessToken(cookie.Value, h.JWTSecret); err == nil {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}
	}

	h.render(w, http.StatusOK, loginPage{Redirect: redirect})
}

func (h *LoginHandler) Submit(w http.ResponseWriter, r *http.Request) {
	// 1. Parse form
	if err := r.ParseForm(); err != nil {
		h.Log.Warn("invalid form", "error", err)
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	redirect := safeRedirect(r.PostFormValue("rd"), h.CookieDomain, h.DefaultRedirect)
	failed := loginPage{Redirect: redirect, Email: email, Error: "Invalid email or password"}

	// 2. Validate required fields
	if email == "" || password == "" {
		h.Log.Warn("missing email or password")
		h.render(w, http.StatusBadRequest, failed)
		return
	}

	// 3. Look up user by email
	var userID, passwordHash string
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT id, email, password_hash FROM users WHERE email = $1`, email,
	).Scan(&userID, &email, &passwordHash)
	if err == sql.ErrNoRows {
		// Timing side-channel mitigation: still compare a dummy hash
		bcrypt.CompareHashAndPassword([]byte("$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), []byte(password))
		h.Log.Info("login failed: unknown email", "email", email)
		h.render(w, http.StatusUnauthorized, failed)
		return
	}
	if err != nil {
		h.Log.Error("db query failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 4. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		h.Log.Info("login failed: wrong password", "user_id", userID)
		h.render(w, http.StatusUnauthorized, failed)
		return
	}

	// 5. Set session cookie shared across subdomains
	if err := setSessionCookie(w, userID, email, h.JWTSecret, h.CookieDomain, h.SessionDuration); err != nil {
		h.Log.Error("failed to generate session token", "error", err, "user_id", userID)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 6. Update last_login_at
	if _, err := h.DB.ExecContext(r.Context(),
		`UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1`, userID,
	); err != nil {
		h.Log.Warn("failed to update last_login_at", "error", err, "user_id", userID)
	}

	// 7. Redirect back to the requested app
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	h.Log.Info("login successful", "user_id", userID)
}

func (h *LoginHandler) render(w http.ResponseWriter, status int, page loginPage) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.Template.Execute(w, page); err != nil {
		h.Log.Error("failed to render login page", "error", err)
	}
}
