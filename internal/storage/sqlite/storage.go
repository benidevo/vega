package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	"github.com/benidevo/vega/internal/job/repository"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
	settingsRepo "github.com/benidevo/vega/internal/settings/repository"
	"github.com/dgraph-io/badger/v4"
	_ "modernc.org/sqlite"
)

// DriveProvider interface for Google Drive operations
type DriveProvider interface {
	Download(ctx context.Context, userID string) (string, error)
	Upload(ctx context.Context, userID string, tempFilePath string) error
	Delete(ctx context.Context, userID string) error
	GetLastModified(ctx context.Context, userID string) (time.Time, error)
}

type Storage struct {
	userID string
	mu     sync.RWMutex

	// Badger for fast in-memory cache
	badgerDB *badger.DB

	// SQLite for persistence
	sqliteDB *sql.DB
	dbPath   string

	// Repositories (for SQLite operations)
	jobRepo      *repository.SQLiteJobRepository
	companyRepo  *repository.SQLiteCompanyRepository
	settingsRepo *settingsRepo.ProfileRepository

	// Sync management
	isDirty    bool
	lastSync   time.Time
	syncTicker *time.Ticker
	stopSync   chan struct{}

	// Google Drive sync (optional)
	driveProvider DriveProvider
	isCloudMode   bool
}

// StorageOptions contains options for creating a new Storage instance
type StorageOptions struct {
	UserID        string
	DataDir       string
	DriveProvider DriveProvider
	IsCloudMode   bool
}

func NewStorage(userID, dataDir string) (*Storage, error) {
	return NewStorageWithOptions(StorageOptions{
		UserID:      userID,
		DataDir:     dataDir,
		IsCloudMode: false,
	})
}

// NewStorageWithOptions creates a new Storage instance with custom options
func NewStorageWithOptions(opts StorageOptions) (*Storage, error) {
	userDir := filepath.Join(opts.DataDir, opts.UserID)
	dbPath := filepath.Join(userDir, "vega.db")
	badgerPath := filepath.Join(userDir, "cache")

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %w", err)
	}

	badgerOpts := badger.DefaultOptions(badgerPath)
	badgerOpts.Logger = nil // Disable badger logging
	badgerDB, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	// In cloud mode, try to download from Google Drive first
	var sqliteDB *sql.DB
	if opts.IsCloudMode && opts.DriveProvider != nil {
		ctx := context.Background()
		tempPath, err := opts.DriveProvider.Download(ctx, opts.UserID)
		if err != nil {
			badgerDB.Close()
			return nil, fmt.Errorf("failed to download from Google Drive: %w", err)
		}

		if tempPath != "" {
			// Copy temp file to local path for SQLite to use
			if err := copyFile(tempPath, dbPath); err != nil {
				badgerDB.Close()
				os.Remove(tempPath)
				return nil, fmt.Errorf("failed to copy temp file: %w", err)
			}
			os.Remove(tempPath)
		}
	}

	sqliteDB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000")
	if err != nil {
		badgerDB.Close()
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure SQLite
	sqliteDB.SetMaxOpenConns(1)
	sqliteDB.SetMaxIdleConns(1)
	if _, err := sqliteDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		badgerDB.Close()
		sqliteDB.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	s := &Storage{
		userID:        opts.UserID,
		badgerDB:      badgerDB,
		sqliteDB:      sqliteDB,
		dbPath:        dbPath,
		lastSync:      time.Now(),
		stopSync:      make(chan struct{}),
		driveProvider: opts.DriveProvider,
		isCloudMode:   opts.IsCloudMode,
	}

	if err := s.migrate(); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	s.companyRepo = repository.NewSQLiteCompanyRepository(sqliteDB)
	s.jobRepo = repository.NewSQLiteJobRepository(sqliteDB, s.companyRepo)
	s.settingsRepo = settingsRepo.NewProfileRepository(sqliteDB)

	// Create user and profile if they don't exist
	userIDInt := hashEmail(opts.UserID)
	if err := s.ensureUserExists(userIDInt, opts.UserID); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to ensure user exists: %w", err)
	}
	if _, err := s.settingsRepo.CreateProfileIfNotExists(context.Background(), userIDInt); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Load data from SQLite to Badger
	if err := s.loadFromSQLite(); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to load data from SQLite: %w", err)
	}

	// Start periodic sync (every 30 seconds)
	// Only start if not in test environment
	if os.Getenv("GO_TEST") != "1" {
		s.startPeriodicSync(30 * time.Second)
	}

	return s, nil
}

// Key prefixes for Badger
const (
	profilePrefix     = "profile:"
	companyPrefix     = "company:"
	jobPrefix         = "job:"
	jobListPrefix     = "jobs:"
	companyListPrefix = "companies:"
)

// GetProfile implements storage.UserStorage
func (s *Storage) GetProfile(ctx context.Context) (*settingsmodels.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var profile settingsmodels.Profile
	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(profilePrefix + s.userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &profile)
		})
	})

	if err == badger.ErrKeyNotFound {
		userIDInt := hashEmail(s.userID)
		dbProfile, err := s.settingsRepo.GetProfileWithRelated(ctx, userIDInt)
		if err != nil {
			return nil, err
		}

		s.cacheProfile(dbProfile)
		return dbProfile, nil
	}

	return &profile, err
}

// SaveProfile implements storage.UserStorage
func (s *Storage) SaveProfile(ctx context.Context, profile *settingsmodels.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure the profile belongs to this user
	userIDInt := hashEmail(s.userID)
	profile.UserID = userIDInt

	// Save to Badger
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	err = s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(profilePrefix+s.userID), data)
	})

	if err == nil {
		s.markDirty()
	}
	return err
}

// ListCompanies implements storage.UserStorage
func (s *Storage) ListCompanies(ctx context.Context) ([]*jobmodels.Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var companies []*jobmodels.Company
	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(companyListPrefix + s.userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &companies)
		})
	})

	if err == badger.ErrKeyNotFound {
		dbCompanies, err := s.companyRepo.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		s.cacheCompanyList(dbCompanies)
		return dbCompanies, nil
	}

	return companies, err
}

// GetCompany implements storage.UserStorage
func (s *Storage) GetCompany(ctx context.Context, companyID int) (*jobmodels.Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var company jobmodels.Company
	key := fmt.Sprintf("%s%d", companyPrefix, companyID)

	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &company)
		})
	})

	if err == badger.ErrKeyNotFound {
		// Load from SQLite
		dbCompany, err := s.companyRepo.GetByID(ctx, companyID)
		if err != nil {
			return nil, err
		}
		// Cache it
		s.cacheCompany(dbCompany)
		return dbCompany, nil
	}

	return &company, err
}

// SaveCompany implements storage.UserStorage
func (s *Storage) SaveCompany(ctx context.Context, company *jobmodels.Company) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// For new companies, generate ID
	if company.ID == 0 {
		company.ID = int(time.Now().UnixNano() % 1000000)
		company.CreatedAt = time.Now()
	}
	company.UpdatedAt = time.Now()

	// Save to Badger
	data, err := json.Marshal(company)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", companyPrefix, company.ID)
	err = s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err == nil {
		s.markDirty()
		s.invalidateCompanyList()
	}
	return err
}

// DeleteCompany implements storage.UserStorage
func (s *Storage) DeleteCompany(ctx context.Context, companyID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s%d", companyPrefix, companyID)
	err := s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err == nil {
		s.markDirty()
		s.invalidateCompanyList()
	}
	return err
}

// ListJobs implements storage.UserStorage
func (s *Storage) ListJobs(ctx context.Context, companyID int) ([]*jobmodels.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*jobmodels.Job
	key := jobListPrefix + s.userID
	if companyID > 0 {
		key = fmt.Sprintf("%s:%d", key, companyID)
	}

	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &jobs)
		})
	})

	if err == badger.ErrKeyNotFound {
		// Load from SQLite
		filter := &jobmodels.JobFilter{}
		if companyID > 0 {
			filter.CompanyID = &companyID
		}
		dbJobs, err := s.jobRepo.GetAll(ctx, *filter)
		if err != nil {
			return nil, err
		}
		// Cache the list
		s.cacheJobList(dbJobs, companyID)
		return dbJobs, nil
	}

	return jobs, err
}

// GetJob implements storage.UserStorage
func (s *Storage) GetJob(ctx context.Context, jobID int) (*jobmodels.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var job jobmodels.Job
	key := fmt.Sprintf("%s%d", jobPrefix, jobID)

	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &job)
		})
	})

	if err == badger.ErrKeyNotFound {
		// Load from SQLite
		dbJob, err := s.jobRepo.GetByID(ctx, jobID)
		if err != nil {
			return nil, err
		}
		// Cache it
		s.cacheJob(dbJob)
		return dbJob, nil
	}

	return &job, err
}

// SaveJob implements storage.UserStorage
func (s *Storage) SaveJob(ctx context.Context, job *jobmodels.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// For new jobs, generate ID
	if job.ID == 0 {
		job.ID = int(time.Now().UnixNano() % 1000000)
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	// Save to Badger
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", jobPrefix, job.ID)
	err = s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err == nil {
		s.markDirty()
		// Invalidate job list cache
		s.invalidateJobList()
	}
	return err
}

// DeleteJob implements storage.UserStorage
func (s *Storage) DeleteJob(ctx context.Context, jobID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s%d", jobPrefix, jobID)
	err := s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err == nil {
		s.markDirty()
		s.invalidateJobList()
	}
	return err
}

// SaveMatchResult implements storage.UserStorage
func (s *Storage) SaveMatchResult(ctx context.Context, result *jobmodels.MatchResult) error {
	// Update the job's match score
	job, err := s.GetJob(ctx, result.JobID)
	if err != nil {
		return err
	}

	job.MatchScore = &result.MatchScore
	return s.SaveJob(ctx, job)
}

// GetMatchHistory implements storage.UserStorage
func (s *Storage) GetMatchHistory(ctx context.Context, limit int) ([]*jobmodels.MatchResult, error) {
	jobs, err := s.ListJobs(ctx, 0)
	if err != nil {
		return nil, err
	}

	userIDInt := hashEmail(s.userID)
	results := make([]*jobmodels.MatchResult, 0)

	for _, job := range jobs {
		if job.MatchScore != nil && *job.MatchScore > 0 {
			results = append(results, &jobmodels.MatchResult{
				ID:         userIDInt*1000000 + job.ID,
				JobID:      job.ID,
				MatchScore: *job.MatchScore,
				Strengths:  []string{},
				Weaknesses: []string{},
				Highlights: []string{},
				Feedback:   "",
				CreatedAt:  job.UpdatedAt,
			})
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// GetMatchResult implements storage.UserStorage
func (s *Storage) GetMatchResult(ctx context.Context, matchID int) (*jobmodels.MatchResult, error) {
	// Extract job ID from match ID
	jobID := matchID % 1000000
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.MatchScore == nil || *job.MatchScore == 0 {
		return nil, fmt.Errorf("no match result found")
	}

	userIDInt := hashEmail(s.userID)
	return &jobmodels.MatchResult{
		ID:         userIDInt*1000000 + job.ID,
		JobID:      job.ID,
		MatchScore: *job.MatchScore,
		Strengths:  []string{},
		Weaknesses: []string{},
		Highlights: []string{},
		Feedback:   "",
		CreatedAt:  job.UpdatedAt,
	}, nil
}

// Sync implements storage.UserStorage
func (s *Storage) Sync(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isDirty {
		return nil
	}

	// Sync from Badger to SQLite
	if err := s.syncToSQLite(ctx); err != nil {
		return err
	}

	// Checkpoint WAL
	_, err := s.sqliteDB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("failed to checkpoint WAL: %w", err)
	}

	s.isDirty = false
	s.lastSync = time.Now()
	return nil
}

// GetLastSyncTime implements storage.UserStorage
func (s *Storage) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSync
}

// Initialize implements storage.UserStorage
func (s *Storage) Initialize(ctx context.Context, userID string) error {
	return nil
}

// Close implements storage.UserStorage
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop periodic sync
	if s.syncTicker != nil {
		s.syncTicker.Stop()
		select {
		case <-s.stopSync:
			// Already closed
		default:
			close(s.stopSync)
		}
		s.syncTicker = nil
	}

	// Final sync
	if s.isDirty {
		s.syncToSQLite(context.Background())
		s.sqliteDB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	}

	var err error
	if s.badgerDB != nil {
		err = s.badgerDB.Close()
		s.badgerDB = nil
	}
	if s.sqliteDB != nil {
		if sqlErr := s.sqliteDB.Close(); sqlErr != nil && err == nil {
			err = sqlErr
		}
		s.sqliteDB = nil
	}

	return err
}

func (s *Storage) markDirty() {
	s.isDirty = true
}

func (s *Storage) startPeriodicSync(interval time.Duration) {
	s.syncTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-s.syncTicker.C:
				s.Sync(context.Background())
			case <-s.stopSync:
				return
			}
		}
	}()
}

func (s *Storage) loadFromSQLite() error {
	ctx := context.Background()
	userIDInt := hashEmail(s.userID)

	if profile, err := s.settingsRepo.GetProfileWithRelated(ctx, userIDInt); err == nil {
		s.cacheProfile(profile)
	}

	if companies, err := s.companyRepo.GetAll(ctx); err == nil {
		for _, company := range companies {
			s.cacheCompany(company)
		}
		s.cacheCompanyList(companies)
	}

	// Load jobs
	if jobs, err := s.jobRepo.GetAll(ctx, jobmodels.JobFilter{}); err == nil {
		for _, job := range jobs {
			s.cacheJob(job)
		}
		s.cacheJobList(jobs, 0)
	}

	return nil
}

func (s *Storage) syncToSQLite(ctx context.Context) error {
	var profile settingsmodels.Profile
	err := s.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(profilePrefix + s.userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &profile)
		})
	})
	if err == nil {
		s.settingsRepo.UpdateProfile(ctx, &profile)
	}

	s.badgerDB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(companyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				var company jobmodels.Company
				if err := json.Unmarshal(val, &company); err == nil {
					// GetOrCreate by name first to ensure it exists
					dbCompany, err := s.companyRepo.GetOrCreate(ctx, company.Name)
					if err == nil && dbCompany != nil {
						// Update with the cached data, preserving the ID from cache
						company.ID = dbCompany.ID // Use DB's ID
						s.companyRepo.Update(ctx, &company)
					}
				}
				return nil
			})
		}
		return nil
	})

	s.badgerDB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(jobPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				var job jobmodels.Job
				if err := json.Unmarshal(val, &job); err == nil {
					if existing, _ := s.jobRepo.GetByID(ctx, job.ID); existing == nil {
						s.jobRepo.Create(ctx, &job)
					} else {
						s.jobRepo.Update(ctx, &job)
					}
				}
				return nil
			})
		}
		return nil
	})

	// If cloud mode is enabled, sync to Google Drive
	if s.isCloudMode && s.driveProvider != nil {
		tempFile, err := os.CreateTemp("", fmt.Sprintf("vega-%s-*.db", s.userID))
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tempPath := tempFile.Name()
		tempFile.Close()

		// Copy SQLite file to temp location
		if err := copyFile(s.dbPath, tempPath); err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("failed to copy database to temp file: %w", err)
		}

		// Upload to Google Drive
		if err := s.driveProvider.Upload(ctx, s.userID, tempPath); err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("failed to upload to Google Drive: %w", err)
		}

		os.Remove(tempPath)
	}

	return nil
}

// Cache helpers
func (s *Storage) cacheProfile(profile *settingsmodels.Profile) {
	if data, err := json.Marshal(profile); err == nil {
		s.badgerDB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(profilePrefix+s.userID), data)
		})
	}
}

func (s *Storage) cacheCompany(company *jobmodels.Company) {
	if data, err := json.Marshal(company); err == nil {
		key := fmt.Sprintf("%s%d", companyPrefix, company.ID)
		s.badgerDB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), data)
		})
	}
}

func (s *Storage) cacheCompanyList(companies []*jobmodels.Company) {
	if data, err := json.Marshal(companies); err == nil {
		s.badgerDB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(companyListPrefix+s.userID), data)
		})
	}
}

func (s *Storage) cacheJob(job *jobmodels.Job) {
	if data, err := json.Marshal(job); err == nil {
		key := fmt.Sprintf("%s%d", jobPrefix, job.ID)
		s.badgerDB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), data)
		})
	}
}

func (s *Storage) cacheJobList(jobs []*jobmodels.Job, companyID int) {
	if data, err := json.Marshal(jobs); err == nil {
		key := jobListPrefix + s.userID
		if companyID > 0 {
			key = fmt.Sprintf("%s:%d", key, companyID)
		}
		s.badgerDB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), data)
		})
	}
}

func (s *Storage) invalidateCompanyList() {
	s.badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(companyListPrefix + s.userID))
	})
}

func (s *Storage) invalidateJobList() {
	s.badgerDB.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(jobListPrefix + s.userID)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			txn.Delete(it.Item().Key())
		}
		return nil
	})
}

// ensureUserExists creates a user record if it doesn't exist
func (s *Storage) ensureUserExists(userID int, email string) error {
	var count int
	err := s.sqliteDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create user for OAuth multi-tenancy
		// Use email as username, dummy password (not used in OAuth), and standard user role (1)
		_, err = s.sqliteDB.Exec(`
			INSERT INTO users (id, username, password, role, created_at, updated_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, userID, email, "oauth-user", 1)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	return nil
}

// hashEmail generates a consistent integer ID from an email address
func hashEmail(email string) int {
	hash := 0
	for _, char := range email {
		hash = ((hash << 5) - hash) + int(char)
		hash = hash & 0x7FFFFFFF // Keep it positive
	}
	if hash == 0 {
		hash = 1
	}
	return hash % 1000000
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// UserDBStatus tracks the status of a user's database
type UserDBStatus struct {
	UserID           string
	HasRemoteDB      bool
	LastSyncTime     time.Time
	MigrationVersion int
	IsNew            bool
}

// CheckUserDBStatus checks if user has existing DB in Google Drive
func (s *Storage) CheckUserDBStatus(ctx context.Context) (*UserDBStatus, error) {
	status := &UserDBStatus{
		UserID: s.userID,
		IsNew:  true,
	}

	if s.isCloudMode && s.driveProvider != nil {
		lastMod, err := s.driveProvider.GetLastModified(ctx, s.userID)
		if err == nil && !lastMod.IsZero() {
			status.HasRemoteDB = true
			status.LastSyncTime = lastMod
			status.IsNew = false
		}
	}

	return status, nil
}

// ApplyMigrationsIfNeeded checks and applies migrations for existing users
func (s *Storage) ApplyMigrationsIfNeeded(ctx context.Context) error {
	// This would check current migration version vs expected
	// For now, migrations are applied in NewStorage, but we could
	// make this more explicit for existing users
	return s.migrate()
}
