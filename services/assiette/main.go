package main

import (
	"database/sql"
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

//go:embed web/index.html
var indexHTML []byte

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

	driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: "assiette_schema_migrations"})
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

	store := &StoreHandler{DB: db, Log: logger}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	r.With(httprate.LimitByIP(5, time.Minute)).Post("/login", store.Login)
	r.With(httprate.LimitByIP(5, time.Minute)).Post("/signup", store.Signup)
	r.Post("/logout", store.Logout)
	r.Route("/api", func(r chi.Router) {
		r.Use(store.WithSession)
		r.Get("/state", store.GetState)
		r.Patch("/state", store.PatchState)
		r.Get("/custom", store.ListCustom)
		r.Post("/custom", store.AddCustom)
		r.Get("/team", store.GetTeam)
		r.Get("/me", store.GetMe)
		r.Put("/me", store.PutMe)
		r.Post("/invites", store.CreateInvite)
		r.Post("/invites/{token}/join", store.JoinTeam)
	})

	logger.Info("assiette service listening", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
