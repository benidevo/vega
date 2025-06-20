package home

import (
	"database/sql"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job/repository"
)

// Setup initializes and returns a new home Handler.
func Setup(db *sql.DB, cfg *config.Settings) *Handler {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo)

	homeService := NewService(jobRepo)

	return NewHandler(cfg, homeService)
}
