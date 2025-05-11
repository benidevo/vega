package dashboard

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job"
)

// Setup initializes and returns a new dashboard Handler.
func Setup(db *sql.DB, cfg *config.Settings) *Handler {
	repo := job.SetupJobRepository(db)
	jobService := job.NewJobService(repo, cfg)
	return NewHandler(cfg, jobService)
}
