package job

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job"
)

// Setup initializes the job API module with its dependencies
func Setup(db *sql.DB, cfg *config.Settings) *JobAPIHandler {
	jobService := job.SetupService(db, cfg)

	return NewJobAPIHandler(jobService)
}
