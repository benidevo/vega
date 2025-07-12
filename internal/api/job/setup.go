package job

import (
	"database/sql"

	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job"
)

// Setup initializes the job API module with its dependencies
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache) *JobAPIHandler {
	jobService := job.SetupService(db, cfg, cache)

	return NewJobAPIHandler(jobService)
}
