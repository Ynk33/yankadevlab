package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: seed <email> <password>")
		os.Exit(1)
	}

	email := strings.ToLower(strings.TrimSpace(os.Args[1]))
	password := os.Args[2]
	if len(password) < 8 {
		log.Fatal("password must be at least 8 characters")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var userID, teamID string
	if err := tx.QueryRow(
		`INSERT INTO assiette.users (email, password_hash) VALUES ($1, $2) RETURNING id`, email, string(hash),
	).Scan(&userID); err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}
	if err := tx.QueryRow(`INSERT INTO assiette.teams DEFAULT VALUES RETURNING id`).Scan(&teamID); err != nil {
		log.Fatalf("failed to create team: %v", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO assiette.team_members (user_id, team_id) VALUES ($1, $2)`, userID, teamID,
	); err != nil {
		log.Fatalf("failed to add team member: %v", err)
	}

	if _, err := tx.Exec(
		`UPDATE assiette.users u SET lang = s.data->>'lang' FROM assiette.state s
		 WHERE u.id = $1 AND s.data->>'lang' IN ('fr', 'en')`, userID,
	); err != nil {
		log.Fatalf("failed to adopt legacy language: %v", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO assiette.team_state (team_id, data, updated_at) SELECT $1, data - 'lang', updated_at FROM assiette.state`, teamID,
	); err != nil {
		log.Fatalf("failed to adopt legacy state: %v", err)
	}
	if _, err := tx.Exec(`DELETE FROM assiette.state`); err != nil {
		log.Fatalf("failed to clear legacy state: %v", err)
	}
	if _, err := tx.Exec(
		`UPDATE assiette.custom_recipes SET team_id = $1 WHERE team_id IS NULL`, teamID,
	); err != nil {
		log.Fatalf("failed to adopt legacy custom recipes: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("failed to commit: %v", err)
	}
	fmt.Printf("user %s seeded with a new team\n", email)
}
