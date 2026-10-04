package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

//go:embed all:web/dist
var webDist embed.FS

const csp = "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:" +
	"; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

// staticHandler serves the Vite build. Hashed files under /assets/ are cached forever, the rest is revalidated.
func staticHandler() http.Handler {
	dist, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServerFS(dist)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			h.Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}

// keyByTraefikRealIP keys rate limits on the client IP. Traefik overwrites X-Real-Ip for untrusted
// clients, so it can't be spoofed as long as the service is only reachable through Traefik.
func keyByTraefikRealIP(r *http.Request) (string, error) {
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip, nil
	}
	return httprate.KeyByIP(r)
}

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
	r.Handle("/*", staticHandler())
	r.With(httprate.Limit(5, time.Minute, httprate.WithKeyFuncs(keyByTraefikRealIP))).Post("/login", store.Login)
	r.With(httprate.Limit(5, time.Minute, httprate.WithKeyFuncs(keyByTraefikRealIP))).Post("/signup", store.Signup)
	r.Post("/logout", store.Logout)
	r.Route("/api", func(r chi.Router) {
		r.Use(store.WithSession)
		r.Get("/state", store.GetState)
		r.Patch("/state", store.PatchState)
		r.Get("/custom", store.ListCustom)
		r.Get("/recipes/{id}/comments", store.ListComments)
		r.Post("/recipes/{id}/comments", store.AddComment)
		r.Put("/comments/{cid}", store.EditComment)
		r.Delete("/comments/{cid}", store.DeleteComment)
		r.Get("/team", store.GetTeam)
		r.Get("/me", store.GetMe)
		r.Put("/me", store.PutMe)
		r.Post("/invites", store.CreateInvite)
		r.Post("/invites/{token}/join", store.JoinTeam)
	})

	logger.Info("assiette service listening", "port", cfg.ServerPort)
	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
