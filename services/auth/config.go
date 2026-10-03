package main

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL         string
	JWTSecret           string
	AccessTokenDuration time.Duration
	SessionDuration     time.Duration
	ServerPort          string
	CookieDomain        string
	LoginURL            string
	DefaultRedirectURL  string
}

func LoadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		return nil, fmt.Errorf("COOKIE_DOMAIN is required")
	}

	loginURL := os.Getenv("LOGIN_URL")
	if loginURL == "" {
		return nil, fmt.Errorf("LOGIN_URL is required")
	}

	defaultRedirectURL := os.Getenv("DEFAULT_REDIRECT_URL")
	if defaultRedirectURL == "" {
		return nil, fmt.Errorf("DEFAULT_REDIRECT_URL is required")
	}

	accessDuration := 15 * time.Minute
	sessionDuration := 7 * 24 * time.Hour

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL:         dbURL,
		JWTSecret:           jwtSecret,
		AccessTokenDuration: accessDuration,
		SessionDuration:     sessionDuration,
		ServerPort:          port,
		CookieDomain:        cookieDomain,
		LoginURL:            loginURL,
		DefaultRedirectURL:  defaultRedirectURL,
	}, nil
}
