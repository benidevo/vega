package job

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job/repository"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings) *JobHandler {
	// Initialize repository
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo)

	// Initialize service
	service := NewJobService(jobRepo, cfg)

	// Create and return handler
	return NewJobHandler(service, cfg)
}
