package badger

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/benidevo/vega/internal/storage"
)

// Provider implements storage.StorageProvider using Badger
type Provider struct {
	basePath string
	caches   map[string]*BadgerCache
	mu       sync.RWMutex
}

// NewProvider creates a new Badger storage provider
func NewProvider(basePath string) *Provider {
	return &Provider{
		basePath: basePath,
		caches:   make(map[string]*BadgerCache),
	}
}

// GetStorage returns a storage instance for the given user
func (p *Provider) GetStorage(ctx context.Context, userID string) (storage.UserStorage, error) {
	p.mu.RLock()
	cache, exists := p.caches[userID]
	p.mu.RUnlock()

	if exists {
		return cache, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// in case another goroutine created it
	if cache, exists = p.caches[userID]; exists {
		return cache, nil
	}

	// user-specific cache directory
	userCachePath := filepath.Join(p.basePath, userID)
	cache, err := NewBadgerCache(userCachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache for user %s: %w", userID, err)
	}

	if err := cache.Initialize(ctx, userID); err != nil {
		cache.Close()
		return nil, fmt.Errorf("failed to initialize cache for user %s: %w", userID, err)
	}

	p.caches[userID] = cache
	return cache, nil
}

// CloseAll closes all storage instances
func (p *Provider) CloseAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for userID, cache := range p.caches {
		if err := cache.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close cache for user %s: %w", userID, err))
		}
		delete(p.caches, userID)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing caches: %v", errs)
	}

	return nil
}
