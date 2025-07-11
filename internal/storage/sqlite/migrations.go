package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// migrate runs all SQL migrations for the user's database
func (s *Storage) migrate() error {
	// Create migrations table if it doesn't exist
	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationsDir, err := s.findMigrationsDir()
	if err != nil {
		return fmt.Errorf("failed to find migrations directory: %w", err)
	}

	migrations, err := s.getMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, migration := range migrations {
		if err := s.applyMigration(migrationsDir, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration, err)
		}
	}

	return nil
}

func (s *Storage) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := s.sqliteDB.Exec(query)
	return err
}

func (s *Storage) findMigrationsDir() (string, error) {
	// Try multiple paths to find migrations (for tests and different working directories)
	possiblePaths := []string{
		filepath.Join("internal", "storage", "sqlite", "migrations"),                   // From project root
		filepath.Join("..", "..", "..", "internal", "storage", "sqlite", "migrations"), // From test directory
		"migrations", // Local to package
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found in any of: %v", possiblePaths)
}

func (s *Storage) getMigrationFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrations = append(migrations, entry.Name())
		}
	}

	// Sort migrations by filename (which should include version numbers)
	sort.Strings(migrations)

	return migrations, nil
}

func (s *Storage) applyMigration(migrationsDir, filename string) error {
	var count int
	err := s.sqliteDB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", filename).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Migration already applied
		return nil
	}

	content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	tx, err := s.sqliteDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}
