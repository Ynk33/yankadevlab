package main

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL     string
	ServerPort      string
	AnthropicAPIKey string
}

func LoadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL:     dbURL,
		ServerPort:      port,
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, nil
}
