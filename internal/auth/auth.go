package auth

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
)

// SetupAuth initializes and returns an AuthHandler using the provided database connection and configuration settings.
// It sets up the user repository, authentication service, and handler dependencies.
func SetupAuth(db *sql.DB, cfg *config.Settings) *AuthHandler {
	repo := NewSQLiteUserRepository(db)
	service := NewAuthService(repo, cfg)
	handler := NewAuthHandler(service)

	return handler
}
