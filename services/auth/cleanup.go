package main

import (
	"database/sql"
	"log/slog"
	"time"
)

func startTokenCleanup(db *sql.DB, logger *slog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		result, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < now() OR revoked_at IS NOT NULL`)
		if err != nil {
			logger.Error("token cleanup failed", "error", err)
			continue
		}

		count, err := result.RowsAffected()
		if err != nil {
			logger.Error("error while reading results", "error", err)
			continue
		}

		logger.Info("token cleanup done", "deleted", count)
	}
}
