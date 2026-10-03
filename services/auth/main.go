package main

import (
	"database/sql"
	_ "embed"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/Ynk33/yankadevlab/services/auth/handler"
)

//go:embed web/login.html
var loginHTML string

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to database")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("failed to create migration driver", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		logger.Error("failed to create migration instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("migrate failed", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	go startTokenCleanup(db, logger, 1*time.Hour)

	loginHandler := &handler.LoginHandler{
		DB:              db,
		Log:             logger,
		JWTSecret:       cfg.JWTSecret,
		SessionDuration: cfg.SessionDuration,
		CookieDomain:    cfg.CookieDomain,
		DefaultRedirect: cfg.DefaultRedirectURL,
		Template:        template.Must(template.New("login").Parse(loginHTML)),
	}

	refreshHandler := &handler.RefreshHandler{
		DB:                   db,
		Log:                  logger,
		JWTSecret:            cfg.JWTSecret,
		AccessTokenDuration:  cfg.AccessTokenDuration,
		RefreshTokenDuration: cfg.SessionDuration,
		CookieDomain:         cfg.CookieDomain,
	}

	logoutHandler := &handler.LogoutHandler{
		Log:          logger,
		CookieDomain: cfg.CookieDomain,
		LoginURL:     cfg.LoginURL,
	}

	verifyHandler := &handler.VerifyHandler{
		Log:             logger,
		JWTSecret:       cfg.JWTSecret,
		LoginURL:        cfg.LoginURL,
		CookieDomain:    cfg.CookieDomain,
		SessionDuration: cfg.SessionDuration,
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://dashboard.yankadevlab.tech"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/login", loginHandler.Show)
	r.With(httprate.LimitByIP(5, time.Minute)).Post("/login", loginHandler.Submit)
	r.Post("/refresh", refreshHandler.ServeHTTP)
	r.Post("/logout", logoutHandler.ServeHTTP)
	r.Get("/verify", verifyHandler.ServeHTTP)

	logger.Info("auth service listening", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
