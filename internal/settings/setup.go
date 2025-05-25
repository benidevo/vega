package settings

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"

	authrepo "github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/settings/repository"
)

// Setup creates a new settings handler and returns it without registering routes
func Setup(cfg *config.Settings, db *sql.DB) *SettingsHandler {
	userRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := repository.NewProfileRepository(db)
	service := NewSettingsService(settingsRepo, cfg, userRepo)
	return NewSettingsHandler(service)
}
