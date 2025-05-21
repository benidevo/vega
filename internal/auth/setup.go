package auth

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/auth/services"
	"github.com/benidevo/ascentio/internal/config"
)

// SetupAuth initializes and returns an AuthHandler using the provided database connection and configuration settings.
// It sets up the user repository, authentication service, and handler dependencies.
func SetupAuth(db *sql.DB, cfg *config.Settings) *AuthHandler {
	repo := repository.NewSQLiteUserRepository(db)
	service := services.NewAuthService(repo, cfg)
	handler := NewAuthHandler(service, cfg)

	return handler
}

// SetupGoogleAuth initializes and returns a GoogleAuthHandler using the provided configuration settings.
// It sets up the GoogleAuthService and handler dependencies.
func SetupGoogleAuth(cfg *config.Settings, db *sql.DB) (*GoogleAuthHandler, error) {
	repo := repository.NewSQLiteUserRepository(db)
	service, err := services.NewGoogleAuthService(cfg, repo)
	if err != nil {
		return nil, err
	}
	handler := NewGoogleAuthHandler(service, cfg)

	return handler, nil
}
