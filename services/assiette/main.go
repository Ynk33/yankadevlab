package main

import (
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
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

var inlineScriptRe = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// indexCSP allows only the inline scripts embedded in index.html, identified by their hash.
var indexCSP = buildCSP(indexHTML)

func buildCSP(html []byte) string {
	var hashes []string
	for _, m := range inlineScriptRe.FindAllSubmatch(html, -1) {
		sum := sha256.Sum256(m[1])
		hashes = append(hashes, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	}
	return "default-src 'self'; script-src " + strings.Join(hashes, " ") +
		"; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com" +
		"; img-src 'self' data:; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'"
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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", indexCSP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Write(indexHTML)
	})
	r.With(httprate.Limit(5, time.Minute, httprate.WithKeyFuncs(keyByTraefikRealIP))).Post("/login", store.Login)
	r.With(httprate.Limit(5, time.Minute, httprate.WithKeyFuncs(keyByTraefikRealIP))).Post("/signup", store.Signup)
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
	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
