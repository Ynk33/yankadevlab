package handler

import (
	"net/http"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
)

const sessionCookieName = "session"

func setSessionCookie(w http.ResponseWriter, userID, email, secret, domain string, duration time.Duration) error {
	sessionToken, err := token.GenerateAccessToken(userID, email, secret, duration)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Domain:   domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(duration),
	})
	return nil
}

func clearSessionCookie(w http.ResponseWriter, domain string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Domain:   domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
