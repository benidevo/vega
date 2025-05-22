package interfaces

import (
	"context"

	"github.com/benidevo/ascentio/internal/job/models"
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
	Update(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status models.JobStatus) error
	GetStats(ctx context.Context) (*models.JobStats, error)
	GetBySourceURL(ctx context.Context, sourceURL string) (*models.Job, error)
	GetOrCreate(ctx context.Context, job *models.Job) (*models.Job, error)
}
