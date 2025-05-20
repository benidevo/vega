package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlDBFromPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("creates database and directory if not exists", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "nested", "path")
		dbPath := filepath.Join(nestedPath, "test.db")

		_, err := os.Stat(nestedPath)
		assert.True(t, os.IsNotExist(err))

		db := SqlDBFromPath(dbPath)
		defer db.Close()

		_, err = os.Stat(nestedPath)
		assert.NoError(t, err)

		_, err = os.Stat(dbPath)
		assert.NoError(t, err)

		err = db.Ping()
		assert.NoError(t, err)
	})

	t.Run("opens existing database", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "existing.db")

		db1 := SqlDBFromPath(dbPath)
		db1.Close()

		db2 := SqlDBFromPath(dbPath)
		defer db2.Close()

		err = db2.Ping()
		assert.NoError(t, err)
	})
}
