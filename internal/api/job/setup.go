package job

import (
	"database/sql"

	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/quota"
)

// Setup initializes the job API module with its dependencies
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache, quotaService *quota.UnifiedService) *JobAPIHandler {
	jobService := job.SetupService(db, cfg, cache)

	return NewJobAPIHandler(jobService, quotaService)
}
