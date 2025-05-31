package job

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job"
)

// Setup initializes the job API module with its dependencies
func Setup(db *sql.DB, cfg *config.Settings) *JobAPIHandler {
	jobRepo := job.SetupJobRepository(db)
	jobService := job.SetupJobService(jobRepo, cfg)

	return NewJobAPIHandler(jobService)
}
