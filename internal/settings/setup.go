package settings

import (
	"context"
	"database/sql"

	"github.com/benidevo/vega/internal/ai"
	authrepo "github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/settings/repository"
)

// Setup creates a new settings handler
func Setup(cfg *config.Settings, db *sql.DB, aiService *ai.AIService, quotaService interface {
	GetAllQuotaStatus(ctx context.Context, userID int) (interface{}, error)
}, authService AuthServiceInterface) *SettingsHandler {
	userRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := repository.NewProfileRepository(db)
	service := NewSettingsService(settingsRepo, cfg, userRepo, authService)
	return NewSettingsHandler(service, aiService, quotaService)
}

// SetupWithService creates a new settings handler and returns both handler and service
func SetupWithService(cfg *config.Settings, db *sql.DB, aiService *ai.AIService, quotaService interface {
	GetAllQuotaStatus(ctx context.Context, userID int) (interface{}, error)
}, authService AuthServiceInterface) (*SettingsHandler, *SettingsService) {
	userRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := repository.NewProfileRepository(db)
	service := NewSettingsService(settingsRepo, cfg, userRepo, authService)
	handler := NewSettingsHandler(service, aiService, quotaService)
	return handler, service
}
