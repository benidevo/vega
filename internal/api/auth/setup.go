package auth

import (
	"database/sql"

	"github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/auth/services"
	"github.com/benidevo/vega/internal/config"
)

// Setup initializes the authentication API handler with the provided database connection and configuration settings.
func Setup(db *sql.DB, cfg *config.Settings) *AuthAPIHandler {
	repo := repository.NewSQLiteUserRepository(db)
	oauthService, _ := services.NewGoogleAuthService(cfg, repo)
	authService := services.NewAuthService(repo, cfg)

	handler := NewAuthAPIHandler(oauthService, authService)

	return handler
}
