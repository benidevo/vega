package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
	"github.com/benidevo/vega/internal/storage"
	"github.com/dgraph-io/badger/v4"
)

// BadgerCache implements storage.UserStorage using Badger as an in-memory cache
type BadgerCache struct {
	db       *badger.DB
	userID   string
	metadata *storage.StorageMetadata
	mu       sync.RWMutex

	// For future phases - Google Drive integration
	driveStorage storage.UserStorage
}

// NewBadgerCache creates a new Badger-based cache
func NewBadgerCache(dbPath string) (*BadgerCache, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable Badger logs for now

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerCache{
		db: db,
		metadata: &storage.StorageMetadata{
			LastSync: time.Now(),
		},
	}, nil
}

// Initialize sets up the cache for a specific user
func (bc *BadgerCache) Initialize(ctx context.Context, userID string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.userID = userID
	bc.metadata.UserID = userID

	// Load metadata if exists
	if err := bc.loadMetadata(); err != nil {
		// Create new metadata if not exists
		bc.metadata = &storage.StorageMetadata{
			UserID:   userID,
			LastSync: time.Now(),
			IsDirty:  false,
		}
		return bc.saveMetadata()
	}

	return nil
}

// Close closes the Badger database
func (bc *BadgerCache) Close() error {
	return bc.db.Close()
}

// Key generation helpers
func (bc *BadgerCache) profileKey() []byte {
	return []byte(fmt.Sprintf("p:%s", bc.userID))
}

func (bc *BadgerCache) companiesKey() []byte {
	return []byte(fmt.Sprintf("c:%s", bc.userID))
}

func (bc *BadgerCache) companyKey(companyID int) []byte {
	return []byte(fmt.Sprintf("c:%s:%d", bc.userID, companyID))
}

func (bc *BadgerCache) jobsKey(companyID int) []byte {
	return []byte(fmt.Sprintf("j:%s:%d", bc.userID, companyID))
}

func (bc *BadgerCache) jobKey(jobID int) []byte {
	return []byte(fmt.Sprintf("j:%s:id:%d", bc.userID, jobID))
}

func (bc *BadgerCache) matchesKey() []byte {
	return []byte(fmt.Sprintf("m:%s", bc.userID))
}

func (bc *BadgerCache) matchKey(matchID int) []byte {
	return []byte(fmt.Sprintf("m:%s:%d", bc.userID, matchID))
}

func (bc *BadgerCache) metadataKey() []byte {
	return []byte(fmt.Sprintf("meta:%s", bc.userID))
}

// Profile operations
func (bc *BadgerCache) GetProfile(ctx context.Context) (*settingsmodels.Profile, error) {
	var profile settingsmodels.Profile

	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.profileKey())
		if err == badger.ErrKeyNotFound {
			// Lazy load from Drive storage if available
			if bc.driveStorage != nil {
				// TODO: Phase 5 - Load from Google Drive
			}
			return storage.ErrProfileNotFound
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &profile)
		})
	})

	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (bc *BadgerCache) SaveProfile(ctx context.Context, profile *settingsmodels.Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bc.profileKey(), data)
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

// Company operations
func (bc *BadgerCache) ListCompanies(ctx context.Context) ([]*jobmodels.Company, error) {
	var companies []*jobmodels.Company

	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.companiesKey())
		if err == badger.ErrKeyNotFound {
			// Return empty list
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &companies)
		})
	})

	if err != nil {
		return nil, err
	}
	return companies, nil
}

func (bc *BadgerCache) GetCompany(ctx context.Context, companyID int) (*jobmodels.Company, error) {
	companies, err := bc.ListCompanies(ctx)
	if err != nil {
		return nil, err
	}

	for _, company := range companies {
		if company.ID == companyID {
			return company, nil
		}
	}

	return nil, storage.ErrCompanyNotFound
}

func (bc *BadgerCache) SaveCompany(ctx context.Context, company *jobmodels.Company) error {
	companies, err := bc.ListCompanies(ctx)
	if err != nil {
		return err
	}

	// Update or append
	found := false
	for i, c := range companies {
		if c.ID == company.ID {
			companies[i] = company
			found = true
			break
		}
	}
	if !found {
		companies = append(companies, company)
	}

	data, err := json.Marshal(companies)
	if err != nil {
		return fmt.Errorf("failed to marshal companies: %w", err)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bc.companiesKey(), data)
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

func (bc *BadgerCache) DeleteCompany(ctx context.Context, companyID int) error {
	companies, err := bc.ListCompanies(ctx)
	if err != nil {
		return err
	}

	// Filter out the company
	filtered := make([]*jobmodels.Company, 0, len(companies))
	for _, c := range companies {
		if c.ID != companyID {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == len(companies) {
		return storage.ErrCompanyNotFound
	}

	data, err := json.Marshal(filtered)
	if err != nil {
		return fmt.Errorf("failed to marshal companies: %w", err)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		// Delete company
		if err := txn.Set(bc.companiesKey(), data); err != nil {
			return err
		}

		// Delete associated jobs
		return txn.Delete(bc.jobsKey(companyID))
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

// Job operations
func (bc *BadgerCache) ListJobs(ctx context.Context, companyID int) ([]*jobmodels.Job, error) {
	var jobs []*jobmodels.Job

	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.jobsKey(companyID))
		if err == badger.ErrKeyNotFound {
			// Return empty list
			jobs = make([]*jobmodels.Job, 0)
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &jobs)
		})
	})

	if err != nil {
		return nil, err
	}

	// Ensure we return an empty slice, not nil
	if jobs == nil {
		jobs = make([]*jobmodels.Job, 0)
	}

	return jobs, nil
}

func (bc *BadgerCache) GetJob(ctx context.Context, jobID int) (*jobmodels.Job, error) {
	// TODO: Optimize with direct job lookup using jobKey()
	// Currently iterates through all companies. It is inefficient for large datasets
	companies, err := bc.ListCompanies(ctx)
	if err != nil {
		return nil, err
	}

	for _, company := range companies {
		jobs, err := bc.ListJobs(ctx, company.ID)
		if err != nil {
			continue
		}

		for _, job := range jobs {
			if job.ID == jobID {
				// Ensure CompanyID is set (in case it wasn't stored)
				job.CompanyID = company.ID
				return job, nil
			}
		}
	}

	return nil, storage.ErrJobNotFound
}

func (bc *BadgerCache) SaveJob(ctx context.Context, job *jobmodels.Job) error {
	jobs, err := bc.ListJobs(ctx, job.CompanyID)
	if err != nil {
		return err
	}

	// Update or append
	found := false
	for i, j := range jobs {
		if j.ID == job.ID {
			jobs[i] = job
			found = true
			break
		}
	}
	if !found {
		jobs = append(jobs, job)
	}

	data, err := json.Marshal(jobs)
	if err != nil {
		return fmt.Errorf("failed to marshal jobs: %w", err)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bc.jobsKey(job.CompanyID), data)
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

func (bc *BadgerCache) DeleteJob(ctx context.Context, jobID int) error {
	// Find the job first to get company ID
	job, err := bc.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	jobs, err := bc.ListJobs(ctx, job.CompanyID)
	if err != nil {
		return err
	}

	// Filter out the job
	filtered := make([]*jobmodels.Job, 0)
	found := false
	for _, j := range jobs {
		if j.ID != jobID {
			filtered = append(filtered, j)
		} else {
			found = true
		}
	}

	if !found {
		// Debug: this shouldn't happen if GetJob succeeded
		// The issue might be that the job IDs don't match
		return fmt.Errorf("job %d not found in company %d jobs", jobID, job.CompanyID)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		key := bc.jobsKey(job.CompanyID)

		if len(filtered) == 0 {
			// Delete the key if no jobs remain
			return txn.Delete(key)
		}

		data, err := json.Marshal(filtered)
		if err != nil {
			return fmt.Errorf("failed to marshal jobs: %w", err)
		}

		return txn.Set(key, data)
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

// Match operations
func (bc *BadgerCache) SaveMatchResult(ctx context.Context, result *jobmodels.MatchResult) error {
	matches, err := bc.GetMatchHistory(ctx, 0)
	if err != nil && err != storage.ErrNoMatches {
		return err
	}

	// Prepend new match
	matches = append([]*jobmodels.MatchResult{result}, matches...)

	data, err := json.Marshal(matches)
	if err != nil {
		return fmt.Errorf("failed to marshal matches: %w", err)
	}

	err = bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bc.matchesKey(), data)
	})

	if err == nil {
		bc.markDirty()
	}

	return err
}

func (bc *BadgerCache) GetMatchHistory(ctx context.Context, limit int) ([]*jobmodels.MatchResult, error) {
	var matches []*jobmodels.MatchResult

	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.matchesKey())
		if err == badger.ErrKeyNotFound {
			return storage.ErrNoMatches
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &matches)
		})
	})

	if err != nil {
		return nil, err
	}

	// Apply limit if specified
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}

func (bc *BadgerCache) GetMatchResult(ctx context.Context, matchID int) (*jobmodels.MatchResult, error) {
	matches, err := bc.GetMatchHistory(ctx, 0)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if match.ID == matchID {
			return match, nil
		}
	}

	return nil, storage.ErrMatchNotFound
}

// Sync operations
func (bc *BadgerCache) Sync(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if !bc.metadata.IsDirty {
		return nil
	}

	// TODO: Phase 5 - Implement Google Drive sync
	// For now, just update sync time
	bc.metadata.LastSync = time.Now()
	bc.metadata.IsDirty = false

	return bc.saveMetadata()
}

func (bc *BadgerCache) GetLastSyncTime() time.Time {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.metadata.LastSync
}

// Helper methods
func (bc *BadgerCache) markDirty() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.metadata.IsDirty = true
}

func (bc *BadgerCache) loadMetadata() error {
	return bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bc.metadataKey())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, bc.metadata)
		})
	})
}

func (bc *BadgerCache) saveMetadata() error {
	data, err := json.Marshal(bc.metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bc.metadataKey(), data)
	})
}
