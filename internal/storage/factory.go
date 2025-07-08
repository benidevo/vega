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
	config      *config.Settings
	db          *sql.DB
	provider    StorageProvider
	driveConfig *DriveProviderConfig
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
	// Check if multi-tenancy is enabled
	if f.config.MultiTenancyEnabled {
		// Note: Badger provider is set externally via SetProvider() in app.go
		// to avoid import cycles. Temporary provider is used as fallback only
		f.provider = &temporaryProvider{db: f.db}
		return nil
	}

	// For now, use temporary provider for non-multi-tenancy mode
	// This will be replaced with SQLite provider in future
	f.provider = &temporaryProvider{db: f.db}

	return nil
}

// GetUserStorage returns a storage instance for the given user
func (f *Factory) GetUserStorage(ctx context.Context, userID string) (UserStorage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	return f.provider.GetStorage(ctx, userID)
}

// GetProvider returns the underlying storage provider
func (f *Factory) GetProvider() StorageProvider {
	return f.provider
}

// SetProvider sets the storage provider (used to avoid import cycles)
func (f *Factory) SetProvider(provider StorageProvider) {
	f.provider = provider
}

// Close closes all storage instances
func (f *Factory) Close() error {
	if f.provider != nil {
		return f.provider.CloseAll()
	}
	return nil
}

// temporaryProvider is a no-op provider used when multi-tenancy is disabled
// In single-tenant mode, the app uses direct SQLite repositories instead of the storage abstraction
// TODO: Future improvement - implement SQLiteProvider to unify storage access patterns
type temporaryProvider struct {
	db *sql.DB
}

func (p *temporaryProvider) GetStorage(ctx context.Context, userID string) (UserStorage, error) {
	// Return a temporary no-op storage for now
	return &temporaryStorage{userID: userID}, nil
}

func (p *temporaryProvider) CloseAll() error {
	return nil
}

// DriveProviderConfig contains configuration for Google Drive storage
type DriveProviderConfig struct {
	CacheProvider StorageProvider
	OAuth2Config  interface{} // *oauth2.Config from golang.org/x/oauth2
}

// SetDriveProvider sets up Google Drive storage provider
func (f *Factory) SetDriveProvider(cfg *DriveProviderConfig) {
	// This will be properly initialized when the drive package is imported
	// For now, we store the config for later use
	f.driveConfig = cfg
}

// GetUserStorageWithToken returns storage for a user with OAuth token
func (f *Factory) GetUserStorageWithToken(ctx context.Context, userID string, token interface{}) (UserStorage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Check if Google Drive is configured
	if f.driveConfig != nil && token != nil {
		// This will be implemented when drive package is imported
		// For now, fall back to regular provider
		return f.provider.GetStorage(ctx, userID)
	}

	return f.provider.GetStorage(ctx, userID)
}

// temporaryStorage is a no-op implementation
// Used when multi-tenancy is disabled since the app uses direct SQLite repositories
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
