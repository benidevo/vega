package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSettings(t *testing.T) {
	settings := NewSettings()

	require.NotNil(t, settings, "Expected config to be initialized")
	assert.Equal(t, ":8080", settings.ServerPort, "Expected default server port to be :8080")
	assert.Equal(t, settings.DBConnectionString, "/app/data/ascentio.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_cache_size=10000&_synchronous=NORMAL", "Expected default DB connection string to be enhanced")
	assert.Equal(t, settings.DBDriver, "sqlite", "Expected default DB driver to be sqlite")
}

func TestGetEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", ":8081")
	t.Setenv("DB_CONNECTION_STRING", "/app/data/test.db?_journal=WAL&_busy_timeout=5000")
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("LOG_LEVEL", "debug")

	t.Run("should return overridden SERVER_PORT from environment variable", func(t *testing.T) {
		value := getEnv("SERVER_PORT", ":8080")
		assert.Equal(t, ":8081", value, "Expected SERVER_PORT to be :8081")
	})

	t.Run("should return overridden DB_CONNECTION_STRING from environment variable", func(t *testing.T) {
		value := getEnv("DB_CONNECTION_STRING", "/app/data/ascentio.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_cache_size=10000&_synchronous=NORMAL")
		assert.Equal(t, "/app/data/test.db?_journal=WAL&_busy_timeout=5000", value, "Expected DB_CONNECTION_STRING to be /app/data/test.db?_journal=WAL&_busy_timeout=5000")
	})

	t.Run("should return overridden DB_DRIVER from environment variable", func(t *testing.T) {
		value := getEnv("DB_DRIVER", "sqlite")
		assert.Equal(t, "postgres", value, "Expected DB_DRIVER to be postgres")
	})

	t.Run("should return overridden LOG_LEVEL from environment variable", func(t *testing.T) {
		value := getEnv("LOG_LEVEL", "info")
		assert.Equal(t, "debug", value, "Expected LOG_LEVEL to be debug")
	})

	t.Run("should use environment variables in NewConfig", func(t *testing.T) {
		settings := NewSettings()
		assert.Equal(t, ":8081", settings.ServerPort)
		assert.Equal(t, "/app/data/test.db?_journal=WAL&_busy_timeout=5000", settings.DBConnectionString)
		assert.Equal(t, "postgres", settings.DBDriver)
		assert.Equal(t, "debug", settings.LogLevel)
	})

	t.Run("should handle empty environment variables", func(t *testing.T) {
		t.Setenv("SERVER_PORT", "")
		value := getEnv("SERVER_PORT", ":8080")
		assert.Equal(t, ":8080", value, "Should use default when env var is empty")
	})
}
