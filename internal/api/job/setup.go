package job

import (
	"database/sql"

	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/monitoring"
)

// Setup initializes the job API module with its dependencies
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache, monitor *monitoring.Monitor) *JobAPIHandler {
	jobService := job.SetupService(db, cfg, cache, monitor)

	return NewJobAPIHandler(jobService)
}
