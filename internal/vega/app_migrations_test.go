package vega

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppMigrations(t *testing.T) {
	t.Run("should_apply_migrations_when_migrations_dir_exists", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "app-migrations-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		migrationsDir := filepath.Join(tempDir, "migrations", "sqlite")
		err = os.MkdirAll(migrationsDir, 0755)
		require.NoError(t, err)

		dbDir := filepath.Join(tempDir, "data")
		err = os.MkdirAll(dbDir, 0755)
		require.NoError(t, err)

		upMigration := `CREATE TABLE app_test_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
		downMigration := `DROP TABLE app_test_users;`

		err = os.WriteFile(filepath.Join(migrationsDir, "000001_create_app_test_users.up.sql"), []byte(upMigration), 0644)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(migrationsDir, "000001_create_app_test_users.down.sql"), []byte(downMigration), 0644)
		require.NoError(t, err)

		cfg, tempDBPath := config.NewTestSettingsWithTempDB()
		cfg.MigrationsDir = migrationsDir
		defer os.Remove(tempDBPath)

		app := New(cfg)
		err = app.Setup()
		require.NoError(t, err)
		defer app.Shutdown(nil)

		err = app.runMigrations()
		require.NoError(t, err)

		var tableName string
		err = app.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='app_test_users'").Scan(&tableName)
		require.NoError(t, err)
		assert.Equal(t, "app_test_users", tableName)

		_, err = app.db.Exec("INSERT INTO app_test_users (username) VALUES (?)", "testuser")
		assert.NoError(t, err)

		var (
			id        int
			username  string
			createdAt sql.NullTime
		)
		err = app.db.QueryRow("SELECT id, username, created_at FROM app_test_users WHERE username = ?", "testuser").Scan(&id, &username, &createdAt)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", username)
		assert.True(t, id > 0)
		assert.True(t, createdAt.Valid)
	})

	t.Run("should_skip_migrations_when_in_test_mode_with_invalid_path", func(t *testing.T) {
		cfg, tempDBPath := config.NewTestSettingsWithTempDB()
		cfg.MigrationsDir = "/nonexistent/path"
		defer os.Remove(tempDBPath)

		app := New(cfg)
		err := app.Setup()
		require.NoError(t, err, "Setup should not fail when IsTest is true, even with invalid migrations path")
		defer app.Shutdown(nil)

		err = app.db.Ping()
		assert.NoError(t, err, "Database should be accessible")
	})
}
