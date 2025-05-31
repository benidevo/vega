package job

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job/interfaces"
	"github.com/benidevo/ascentio/internal/job/repository"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings) *JobHandler {
	jobRepo := SetupJobRepository(db)
	service := SetupJobService(jobRepo, cfg)

	return NewJobHandler(service, cfg)
}

// SetupJobRepository initializes and returns a job repository.
func SetupJobRepository(db *sql.DB) interfaces.JobRepository {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	return repository.NewSQLiteJobRepository(db, companyRepo)
}

// SetupJobService initializes and returns a new JobService using the provided JobRepository and configuration settings.
func SetupJobService(repo interfaces.JobRepository, cfg *config.Settings) *JobService {
	return NewJobService(repo, cfg)
}
