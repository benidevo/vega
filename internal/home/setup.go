package home

import (
	"database/sql"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job/repository"
)

// Setup initializes and returns a new home Handler.
func Setup(db *sql.DB, cfg *config.Settings) *Handler {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo)

	homeService := NewService(jobRepo)

	return NewHandler(cfg, homeService)
}
