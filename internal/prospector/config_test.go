package prospector

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	require.NotNil(t, config, "Expected config to be initialized")
	assert.Equal(t, config.ServerPort, ":8080", "Expected default server port to be :8080")
	assert.Equal(t, config.DBConnectionString, "/app/data/prospector.db?_journal=WAL&_busy_timeout=5000", "Expected default DB connection string to be /app/data/prospector.db?_journal=WAL&_busy_timeout=5000")
	assert.Equal(t, config.DBDriver, "sqlite", "Expected default DB driver to be sqlite")
}

func TestGetEnv(t *testing.T) {
	t.Setenv("SERVER_PORT", ":8081")
	t.Setenv("DB_CONNECTION_STRING", "/app/data/test.db?_journal=WAL&_busy_timeout=5000")
	t.Setenv("DB_DRIVER", "postgres")

	t.Run("should return overridden SERVER_PORT from environment variable", func(t *testing.T) {
		value := getEnv("SERVER_PORT", ":8080")
		assert.Equal(t, ":8081", value, "Expected SERVER_PORT to be :8081")
	})

	t.Run("should return overridden DB_CONNECTION_STRING from environment variable", func(t *testing.T) {
		value := getEnv("DB_CONNECTION_STRING", "/app/data/prospector.db?_journal=WAL&_busy_timeout=5000")
		assert.Equal(t, "/app/data/test.db?_journal=WAL&_busy_timeout=5000", value, "Expected DB_CONNECTION_STRING to be /app/data/test.db?_journal=WAL&_busy_timeout=5000")
	})

	t.Run("should return overridden DB_DRIVER from environment variable", func(t *testing.T) {
		value := getEnv("DB_DRIVER", "sqlite")
		assert.Equal(t, "postgres", value, "Expected DB_DRIVER to be postgres")
	})

	t.Run("should use environment variables in NewConfig", func(t *testing.T) {
		config := NewConfig()
		assert.Equal(t, ":8081", config.ServerPort)
		assert.Equal(t, "/app/data/test.db?_journal=WAL&_busy_timeout=5000", config.DBConnectionString)
		assert.Equal(t, "postgres", config.DBDriver)
	})

	t.Run("should handle empty environment variables", func(t *testing.T) {
		t.Setenv("SERVER_PORT", "")
		value := getEnv("SERVER_PORT", ":8080")
		assert.Equal(t, ":8080", value, "Should use default when env var is empty")
	})
}
