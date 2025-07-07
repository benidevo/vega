package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/benidevo/vega/internal/config"
	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
)

// Factory creates storage instances based on configuration
type Factory struct {
	config   *config.Settings
	db       *sql.DB
	provider StorageProvider
}

// NewFactory creates a new storage factory
func NewFactory(cfg *config.Settings, db *sql.DB) (*Factory, error) {
	f := &Factory{
		config: cfg,
		db:     db,
	}

	if err := f.initializeProvider(); err != nil {
		return nil, err
	}

	return f, nil
}

// initializeProvider sets up the storage provider based on configuration
func (f *Factory) initializeProvider() error {
	// For now, always use SQLite storage
	// In future phases, this will check for GoogleDriveStorage flag
	// and initialize appropriate provider

	// For Phase 2, we'll use a temporary provider
	// Phase 5 will implement the actual providers
	f.provider = &temporaryProvider{db: f.db}

	return nil
}

// GetUserStorage returns a storage instance for the given user
func (f *Factory) GetUserStorage(userID string) (UserStorage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	return f.provider.GetStorage(userID)
}

// GetProvider returns the underlying storage provider
func (f *Factory) GetProvider() StorageProvider {
	return f.provider
}

// Close closes all storage instances
func (f *Factory) Close() error {
	if f.provider != nil {
		return f.provider.CloseAll()
	}
	return nil
}

// temporaryProvider is a placeholder for Phase 2
type temporaryProvider struct {
	db *sql.DB
}

func (p *temporaryProvider) GetStorage(userID string) (UserStorage, error) {
	// Return a temporary no-op storage for now
	return &temporaryStorage{userID: userID}, nil
}

func (p *temporaryProvider) CloseAll() error {
	return nil
}

// temporaryStorage is a no-op implementation for Phase 2
type temporaryStorage struct {
	userID string
}

func (s *temporaryStorage) GetProfile(ctx context.Context) (*settingsmodels.Profile, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) SaveProfile(ctx context.Context, profile *settingsmodels.Profile) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) ListCompanies(ctx context.Context) ([]*jobmodels.Company, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) GetCompany(ctx context.Context, companyID int) (*jobmodels.Company, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) SaveCompany(ctx context.Context, company *jobmodels.Company) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) DeleteCompany(ctx context.Context, companyID int) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) ListJobs(ctx context.Context, companyID int) ([]*jobmodels.Job, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) GetJob(ctx context.Context, jobID int) (*jobmodels.Job, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) SaveJob(ctx context.Context, job *jobmodels.Job) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) DeleteJob(ctx context.Context, jobID int) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) SaveMatchResult(ctx context.Context, result *jobmodels.MatchResult) error {
	return fmt.Errorf("not implemented")
}

func (s *temporaryStorage) GetMatchHistory(ctx context.Context, limit int) ([]*jobmodels.MatchResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) GetMatchResult(ctx context.Context, matchID int) (*jobmodels.MatchResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *temporaryStorage) Sync(ctx context.Context) error {
	return nil
}

func (s *temporaryStorage) GetLastSyncTime() time.Time {
	return time.Now()
}

func (s *temporaryStorage) Initialize(ctx context.Context, userID string) error {
	return nil
}

func (s *temporaryStorage) Close() error {
	return nil
}
