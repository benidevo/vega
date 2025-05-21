package job

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job/interfaces"
	"github.com/benidevo/ascentio/internal/job/repository"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings) *JobHandler {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo)
	service := NewJobService(jobRepo, cfg)

	return NewJobHandler(service, cfg)
}

// SetupJobRepository initializes and returns a job repository.
func SetupJobRepository(db *sql.DB) interfaces.JobRepository {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	return repository.NewSQLiteJobRepository(db, companyRepo)
}
