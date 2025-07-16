package home

import (
	"database/sql"

	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/job/repository"
)

// Setup initializes and returns a new home Handler.
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache, jobService *job.JobService) *Handler {
	homeService := SetupService(db, cache, jobService)

	return NewHandler(cfg, homeService)
}

// SetupService initializes just the home service without the handler.
func SetupService(db *sql.DB, cache cache.Cache, jobService *job.JobService) *Service {
	companyRepo := repository.NewSQLiteCompanyRepository(db, cache)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo, cache)

	return NewService(jobRepo, jobService)
}
