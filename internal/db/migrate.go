package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/benidevo/prospector/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// MigrateDatabase performs database migrations using the golang-migrate library.
// It applies all up migrations from the specified migrations directory to the database.
func MigrateDatabase(dbPath, migrationsDir string) error {
	log.Info().
		Str("dbPath", dbPath).
		Str("migrationsDir", migrationsDir).
		Msg("Starting database migration")

	driver, err := sqlite.WithInstance(logger.SqlDBFromPath(dbPath), &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver instance: %w", err)
	}

	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", absPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite",
		driver,
	)

	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	log.Info().Msg("Running database migrations...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("No migrations to apply, database is up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Database migration completed successfully")
	return nil
}
