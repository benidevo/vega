package job

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/ai"
	authrepo "github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job/interfaces"
	"github.com/benidevo/ascentio/internal/job/repository"
	"github.com/benidevo/ascentio/internal/settings"
	settingsrepo "github.com/benidevo/ascentio/internal/settings/repository"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings) *JobHandler {
	service := SetupService(db, cfg)
	return NewJobHandler(service, cfg)
}

// SetupService initializes just the job service without the handler.
func SetupService(db *sql.DB, cfg *config.Settings) *JobService {
	jobRepo := SetupJobRepository(db)
	aiService, err := SetupAIService(cfg)
	if err != nil {
		// AI service is optional.
		// When nil, AI-dependent features (job matching, cover letter generation) will return
		// ErrAIServiceUnavailable.
		aiService = nil
	}

	settingsService := SetupSettingsService(db, cfg)
	return SetupJobService(jobRepo, aiService, settingsService, cfg)
}

// SetupJobRepository initializes and returns a job repository.
func SetupJobRepository(db *sql.DB) interfaces.JobRepository {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	return repository.NewSQLiteJobRepository(db, companyRepo)
}

// SetupAIService initializes and returns an AI service instance.
func SetupAIService(cfg *config.Settings) (*ai.AIService, error) {
	return ai.Setup(cfg)
}

// SetupSettingsService initializes and returns a settings service instance.
func SetupSettingsService(db *sql.DB, cfg *config.Settings) *settings.SettingsService {
	// Create the service directly instead of using Setup which returns handler
	authRepo := authrepo.NewSQLiteUserRepository(db)
	settingsRepo := settingsrepo.NewProfileRepository(db)
	return settings.NewSettingsService(settingsRepo, cfg, authRepo)
}

// SetupJobService initializes and returns a new JobService using the provided JobRepository, AIService, SettingsService and configuration settings.
func SetupJobService(repo interfaces.JobRepository, aiService *ai.AIService, settingsService *settings.SettingsService, cfg *config.Settings) *JobService {
	return NewJobService(repo, aiService, settingsService, cfg)
}
