package db

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// SqlDBFromPath creates a new SQL database connection from a path.
func SqlDBFromPath(dbPath string) *sql.DB {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal().Err(err).Str("path", dir).Msg("Failed to create database directory")
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to open database connection")
	}

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to ping database")
	}

	return db
}
