package interfaces

import (
	"context"

	"github.com/benidevo/vega/internal/job/models"
)

// CompanyRepository defines methods for interacting with company data
type CompanyRepository interface {
	GetOrCreate(ctx context.Context, name string) (*models.Company, error)
	GetByID(ctx context.Context, id int) (*models.Company, error)
	GetByName(ctx context.Context, name string) (*models.Company, error)
	GetAll(ctx context.Context) ([]*models.Company, error)
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, company *models.Company) error
}

// JobRepository defines methods for interacting with job data
type JobRepository interface {
	Create(ctx context.Context, job *models.Job) (*models.Job, error)
	GetByID(ctx context.Context, id int) (*models.Job, error)
	GetAll(ctx context.Context, filter models.JobFilter) ([]*models.Job, error)
	GetCount(ctx context.Context, filter models.JobFilter) (int, error)
	Update(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status models.JobStatus) error
	UpdateMatchScore(ctx context.Context, jobID int, matchScore *int) error
	GetStats(ctx context.Context) (*models.JobStats, error)
	GetBySourceURL(ctx context.Context, sourceURL string) (*models.Job, error)
	GetOrCreate(ctx context.Context, job *models.Job) (*models.Job, error)

	CreateMatchResult(ctx context.Context, matchResult *models.MatchResult) error
	GetJobMatchHistory(ctx context.Context, jobID int) ([]*models.MatchResult, error)
	GetRecentMatchResults(ctx context.Context, limit int) ([]*models.MatchResult, error)
	GetRecentMatchResultsWithDetails(ctx context.Context, limit int, currentJobID int) ([]*models.MatchSummary, error)
	DeleteMatchResult(ctx context.Context, matchID int) error
	MatchResultBelongsToJob(ctx context.Context, matchID, jobID int) (bool, error)
}
