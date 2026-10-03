package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ynk33/yankadevlab/services/auth/token"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func mustToken(t *testing.T, secret string, duration time.Duration) string {
	t.Helper()
	tok, err := token.GenerateAccessToken("user-1", "me@example.com", secret, duration)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func newTestVerifyHandler() *VerifyHandler {
	return &VerifyHandler{
		Log:             slog.New(slog.NewTextHandler(io.Discard, nil)),
		JWTSecret:       testSecret,
		LoginURL:        "https://dashboard.example.tech/login",
		CookieDomain:    ".example.tech",
		SessionDuration: 7 * 24 * time.Hour,
	}
}

func TestVerifyHandler(t *testing.T) {
	valid := mustToken(t, testSecret, time.Hour)
	expired := mustToken(t, testSecret, -time.Hour)
	forged := mustToken(t, "other-secret", time.Hour)

	tests := []struct {
		name         string
		cookie       string
		accept       string
		wantStatus   int
		wantLocation string
		wantUserID   string
	}{
		{name: "valid cookie", cookie: valid, wantStatus: http.StatusOK, wantUserID: "user-1"},
		{name: "expired cookie api", cookie: expired, accept: "application/json", wantStatus: http.StatusUnauthorized},
		{name: "forged cookie api", cookie: forged, wantStatus: http.StatusUnauthorized},
		{
			name:         "expired cookie navigation",
			cookie:       expired,
			accept:       "text/html,application/xhtml+xml",
			wantStatus:   http.StatusFound,
			wantLocation: "https://dashboard.example.tech/login?rd=https%3A%2F%2Fapp.example.tech%2Fpath%3Fq%3D1",
		},
		{
			name:         "no credentials navigation",
			accept:       "text/html",
			wantStatus:   http.StatusFound,
			wantLocation: "https://dashboard.example.tech/login?rd=https%3A%2F%2Fapp.example.tech%2Fpath%3Fq%3D1",
		},
		{name: "no credentials api", accept: "application/json", wantStatus: http.StatusUnauthorized},
	}

	h := newTestVerifyHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/verify", nil)
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Host", "app.example.tech")
			req.Header.Set("X-Forwarded-Uri", "/path?q=1")
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tt.cookie})
			}
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("Location = %q, want %q", got, tt.wantLocation)
			}
			if got := rec.Header().Get("X-User-Id"); got != tt.wantUserID {
				t.Errorf("X-User-Id = %q, want %q", got, tt.wantUserID)
			}
		})
	}
}

func TestVerifySessionRenewal(t *testing.T) {
	issuedAgo := func(d time.Duration) string {
		now := time.Now()
		claims := token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-1",
				IssuedAt:  jwt.NewNumericDate(now.Add(-d)),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			},
			Email: "me@example.com",
		}
		tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}

	tests := []struct {
		name      string
		cookie    string
		wantRenew bool
	}{
		{name: "fresh session", cookie: issuedAgo(time.Hour), wantRenew: false},
		{name: "stale session", cookie: issuedAgo(48 * time.Hour), wantRenew: true},
	}

	h := newTestVerifyHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/verify", nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tt.cookie})

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			renewed := false
			for _, c := range rec.Result().Cookies() {
				if c.Name == sessionCookieName && c.Value != "" {
					renewed = true
				}
			}
			if renewed != tt.wantRenew {
				t.Errorf("renewed = %v, want %v", renewed, tt.wantRenew)
			}
		})
	}
}
