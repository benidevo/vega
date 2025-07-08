package drive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/benidevo/vega/internal/storage"
)

const (
	defaultSyncInterval = 5 * time.Minute
	maxRetryAttempts    = 3
	initialBackoff      = 1 * time.Second
	maxBackoff          = 30 * time.Second
)

// SyncManager handles automatic synchronization with debouncing
type SyncManager struct {
	storage  *DriveStorage
	interval time.Duration
	timer    *time.Timer
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	syncChan chan struct{}
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewSyncManager creates a new sync manager
func NewSyncManager(storage *DriveStorage, interval time.Duration) *SyncManager {
	if interval <= 0 {
		interval = defaultSyncInterval
	}

	ctx, cancel := context.WithCancel(context.Background())

	sm := &SyncManager{
		storage:  storage,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		syncChan: make(chan struct{}, 1),
		stopChan: make(chan struct{}),
	}

	return sm
}

// Start begins the sync manager
func (sm *SyncManager) Start() {
	sm.wg.Add(1)
	go sm.syncLoop()
}

// Stop gracefully shuts down the sync manager
func (sm *SyncManager) Stop() {
	close(sm.stopChan)
	sm.cancel()
	sm.wg.Wait()

	// Final sync
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = sm.syncWithRetry(ctx)
}

// TriggerSync triggers a debounced sync
func (sm *SyncManager) TriggerSync() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Cancel existing timer
	if sm.timer != nil {
		sm.timer.Stop()
	}

	// Set new timer
	sm.timer = time.AfterFunc(sm.interval, func() {
		select {
		case sm.syncChan <- struct{}{}:
		default:
			// Channel full, sync already pending
		}
	})
}

// syncLoop is the main sync loop
func (sm *SyncManager) syncLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			if sm.storage.needsSync() {
				_ = sm.syncWithRetry(sm.ctx)
			}
		case <-sm.syncChan:
			_ = sm.syncWithRetry(sm.ctx)
		}
	}
}

// syncWithRetry performs sync with exponential backoff retry
func (sm *SyncManager) syncWithRetry(ctx context.Context) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			// Exponential backoff
			backoff = time.Duration(math.Min(float64(backoff*2), float64(maxBackoff)))
		}

		err := sm.storage.Sync(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		fmt.Printf("Sync attempt %d failed: %v\n", attempt+1, err)
	}

	return fmt.Errorf("sync failed after %d attempts: %w", maxRetryAttempts, lastErr)
}

// ConflictResolution handles conflicts during sync
type ConflictResolution int

const (
	ConflictResolutionLastWriteWins ConflictResolution = iota
	ConflictResolutionMerge
	ConflictResolutionManual
)

// ResolveConflict handles sync conflicts
func ResolveConflict(local, remote *storage.UserDocument, resolution ConflictResolution) (*storage.UserDocument, error) {
	switch resolution {
	case ConflictResolutionLastWriteWins:
		// Choose the document with the latest timestamp
		if local.UpdatedAt.After(remote.UpdatedAt) {
			return local, nil
		}
		return remote, nil

	case ConflictResolutionMerge:
		// Merge strategy: combine data from both documents
		// This is a simple implementation - could be more sophisticated
		merged := &storage.UserDocument{
			UpdatedAt: time.Now(),
			Data:      storage.UserDataCore{},
		}

		// Use latest profile
		if local.UpdatedAt.After(remote.UpdatedAt) {
			merged.Data.Profile = local.Data.Profile
		} else {
			merged.Data.Profile = remote.Data.Profile
		}

		// Merge companies and jobs by ID
		companyMap := make(map[int]*storage.Company)
		jobMap := make(map[int]*storage.Job)
		matchMap := make(map[int]*storage.MatchResult)

		// Add all companies
		for _, c := range local.Data.Companies {
			companyMap[c.ID] = c
		}
		for _, c := range remote.Data.Companies {
			if existing, exists := companyMap[c.ID]; !exists || c.UpdatedAt.After(existing.UpdatedAt) {
				companyMap[c.ID] = c
			}
		}

		// Add all jobs
		for _, j := range local.Data.Jobs {
			jobMap[j.ID] = j
		}
		for _, j := range remote.Data.Jobs {
			if existing, exists := jobMap[j.ID]; !exists || j.UpdatedAt.After(existing.UpdatedAt) {
				jobMap[j.ID] = j
			}
		}

		// Add all matches
		for _, m := range local.Data.Matches {
			matchMap[m.ID] = m
		}
		for _, m := range remote.Data.Matches {
			if existing, exists := matchMap[m.ID]; !exists || m.CreatedAt.After(existing.CreatedAt) {
				matchMap[m.ID] = m
			}
		}

		// Convert maps back to slices
		for _, c := range companyMap {
			merged.Data.Companies = append(merged.Data.Companies, c)
		}
		for _, j := range jobMap {
			merged.Data.Jobs = append(merged.Data.Jobs, j)
		}
		for _, m := range matchMap {
			merged.Data.Matches = append(merged.Data.Matches, m)
		}

		merged.UpdateChecksum()
		return merged, nil

	case ConflictResolutionManual:
		return nil, fmt.Errorf("manual conflict resolution not implemented")

	default:
		return nil, fmt.Errorf("unknown conflict resolution strategy")
	}
}
