package storage

import (
	"context"
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
)

// UserStorage defines the interface for user-specific data storage
type UserStorage interface {
	GetProfile(ctx context.Context) (*settingsmodels.Profile, error)
	SaveProfile(ctx context.Context, profile *settingsmodels.Profile) error

	ListCompanies(ctx context.Context) ([]*jobmodels.Company, error)
	GetCompany(ctx context.Context, companyID int) (*jobmodels.Company, error)
	SaveCompany(ctx context.Context, company *jobmodels.Company) error
	DeleteCompany(ctx context.Context, companyID int) error

	ListJobs(ctx context.Context, companyID int) ([]*jobmodels.Job, error)
	GetJob(ctx context.Context, jobID int) (*jobmodels.Job, error)
	SaveJob(ctx context.Context, job *jobmodels.Job) error
	DeleteJob(ctx context.Context, jobID int) error

	SaveMatchResult(ctx context.Context, result *jobmodels.MatchResult) error
	GetMatchHistory(ctx context.Context, limit int) ([]*jobmodels.MatchResult, error)
	GetMatchResult(ctx context.Context, matchID int) (*jobmodels.MatchResult, error)

	Sync(ctx context.Context) error
	GetLastSyncTime() time.Time

	Initialize(ctx context.Context, userID string) error
	Close() error
}

// StorageProvider creates storage instances for users
type StorageProvider interface {
	// GetStorage returns a storage instance for the given user
	GetStorage(ctx context.Context, userID string) (UserStorage, error)

	// CloseAll closes all storage instances
	CloseAll() error
}
