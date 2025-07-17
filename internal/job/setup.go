package job

import (
	"database/sql"

	"github.com/benidevo/vega/internal/ai"
	authrepo "github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job/interfaces"
	"github.com/benidevo/vega/internal/job/repository"
	"github.com/benidevo/vega/internal/monitoring"
	"github.com/benidevo/vega/internal/quota"
	"github.com/benidevo/vega/internal/settings"
	settingsrepo "github.com/benidevo/vega/internal/settings/repository"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache, monitor *monitoring.Monitor) *JobHandler {
	service := SetupService(db, cfg, cache, monitor)
	return NewJobHandler(service, cfg)
}

// SetupService initializes just the job service without the handler.
func SetupService(db *sql.DB, cfg *config.Settings, cache cache.Cache, monitor *monitoring.Monitor) *JobService {
	jobRepo := SetupJobRepository(db, cache)
	aiService, err := SetupAIService(cfg)
	if err != nil {
		// AI service is optional.
		// When nil, AI-dependent features (job matching, cover letter generation) will return
		// ErrAIServiceUnavailable.
		aiService = nil
	}

	settingsService := SetupSettingsService(db, cfg)

	// Setup quota service
	quotaAdapter := quota.NewJobRepositoryAdapter(jobRepo)
	quotaService := quota.NewService(db, quotaAdapter, cfg.IsCloudMode)

	// Wrap quota service with monitoring if available and in cloud mode
	var quotaServiceWithMetrics quota.QuotaService
	if monitor != nil && cfg.IsCloudMode {
		quotaServiceWithMetrics = monitoring.NewQuotaServiceWithMetrics(quotaService, monitor)
	} else {
		quotaServiceWithMetrics = quotaService
	}

	jobService := NewJobService(jobRepo, aiService, settingsService, quotaServiceWithMetrics, cfg)
	if monitor != nil && cfg.IsCloudMode {
		jobService.monitor = monitor
	}

	return jobService
}

// SetupJobRepository initializes and returns a job repository.
func SetupJobRepository(db *sql.DB, cache cache.Cache) interfaces.JobRepository {
	companyRepo := repository.NewSQLiteCompanyRepository(db, cache)
	return repository.NewSQLiteJobRepository(db, companyRepo, cache)
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

