package main

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	ServerPort           string
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

	accessDuration := 15 * time.Minute
	refreshDuration := 7 * 24 * time.Hour

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL:          dbURL,
		JWTSecret:            jwtSecret,
		AccessTokenDuration:  accessDuration,
		RefreshTokenDuration: refreshDuration,
		ServerPort:           port,
	}, nil
}
