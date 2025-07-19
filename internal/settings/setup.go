package settings

import (
	"database/sql"

	"github.com/benidevo/vega/internal/ai"
	authrepo "github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/settings/repository"
)

// Setup creates a new settings handler
func Setup(cfg *config.Settings, db *sql.DB, aiService *ai.AIService) *SettingsHandler {
	userRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := repository.NewProfileRepository(db)
	service := NewSettingsService(settingsRepo, cfg, userRepo)
	return NewSettingsHandler(service, aiService)
}

// SetupWithService creates a new settings handler and returns both handler and service
func SetupWithService(cfg *config.Settings, db *sql.DB, aiService *ai.AIService) (*SettingsHandler, *SettingsService) {
	userRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := repository.NewProfileRepository(db)
	service := NewSettingsService(settingsRepo, cfg, userRepo)
	handler := NewSettingsHandler(service, aiService)
	return handler, service
}
