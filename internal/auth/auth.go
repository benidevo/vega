package auth

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
)

func SetupAuth(db *sql.DB, cfg *config.Settings) *AuthHandler {
	repo := NewSQLiteUserRepository(db)
	service := NewAuthService(repo, cfg)
	handler := NewAuthHandler(service)

	return handler
}
