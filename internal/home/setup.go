package home

import (
	"database/sql"

	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job/repository"
)

// Setup initializes and returns a new home Handler.
func Setup(db *sql.DB, cfg *config.Settings, cache cache.Cache) *Handler {
	companyRepo := repository.NewSQLiteCompanyRepository(db, cache)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo, cache)

	homeService := NewService(jobRepo)

	return NewHandler(cfg, homeService)
}
