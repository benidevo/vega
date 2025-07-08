package drive

import (
	"context"
	"fmt"
	"sync"

	"github.com/benidevo/vega/internal/storage"
	"golang.org/x/oauth2"
)

// Provider manages Google Drive storage instances for multiple users
type Provider struct {
	cacheProvider storage.StorageProvider
	config        *oauth2.Config
	storages      map[string]*DriveStorage
	syncManagers  map[string]*SyncManager
	mu            sync.RWMutex
}

// NewProvider creates a new Google Drive storage provider
func NewProvider(cacheProvider storage.StorageProvider, config *oauth2.Config) *Provider {
	return &Provider{
		cacheProvider: cacheProvider,
		config:        config,
		storages:      make(map[string]*DriveStorage),
		syncManagers:  make(map[string]*SyncManager),
	}
}

// GetStorage returns a storage instance for the given user
func (p *Provider) GetStorage(ctx context.Context, userID string, token *oauth2.Token) (storage.UserStorage, error) {
	p.mu.RLock()
	if ds, exists := p.storages[userID]; exists {
		p.mu.RUnlock()
		return ds, nil
	}
	p.mu.RUnlock()

	// Create new instance
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if ds, exists := p.storages[userID]; exists {
		return ds, nil
	}

	// Get cache storage
	cache, err := p.cacheProvider.GetStorage(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache storage: %w", err)
	}

	// Create Drive storage
	ds, err := NewDriveStorage(ctx, userID, token, p.config, cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create drive storage: %w", err)
	}

	// Create and start sync manager
	sm := NewSyncManager(ds, 0) // Use default interval
	sm.Start()

	p.storages[userID] = ds
	p.syncManagers[userID] = sm

	return ds, nil
}

// Close shuts down all storage instances
func (p *Provider) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop all sync managers
	for userID, sm := range p.syncManagers {
		sm.Stop()
		delete(p.syncManagers, userID)
	}

	// Close all storages
	var errs []error
	for userID, ds := range p.storages {
		if err := ds.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close storage for user %s: %w", userID, err))
		}
		delete(p.storages, userID)
	}

	// Close cache provider
	if err := p.cacheProvider.CloseAll(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close cache provider: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// TriggerSync triggers a sync for a specific user
func (p *Provider) TriggerSync(userID string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sm, exists := p.syncManagers[userID]
	if !exists {
		return fmt.Errorf("no sync manager for user %s", userID)
	}

	sm.TriggerSync()
	return nil
}
