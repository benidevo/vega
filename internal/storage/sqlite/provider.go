package sqlite

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benidevo/vega/internal/storage"
)

// Provider manages SQLite storage instances for multiple users
type Provider struct {
	dataDir       string
	instances     map[string]*Storage
	mu            sync.RWMutex
	driveProvider storage.DriveProvider
	isCloudMode   bool
}

// NewProvider creates a new SQLite storage provider
func NewProvider(dataDir string) *Provider {
	return &Provider{
		dataDir:   dataDir,
		instances: make(map[string]*Storage),
	}
}

// NewProviderWithDrive creates a new SQLite storage provider with Google Drive support
func NewProviderWithDrive(dataDir string, driveProvider storage.DriveProvider) *Provider {
	return &Provider{
		dataDir:       dataDir,
		instances:     make(map[string]*Storage),
		driveProvider: driveProvider,
		isCloudMode:   true,
	}
}

// GetStorage retrieves or creates a storage instance for a user
func (p *Provider) GetStorage(ctx context.Context, userID string) (storage.UserStorage, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	p.mu.RLock()
	instance, exists := p.instances[userID]
	p.mu.RUnlock()

	if exists {
		return instance, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if instance, exists := p.instances[userID]; exists {
		return instance, nil
	}

	var newInstance *Storage
	var err error

	if p.isCloudMode && p.driveProvider != nil {
		newInstance, err = NewStorageWithOptions(StorageOptions{
			UserID:        userID,
			DataDir:       p.dataDir,
			DriveProvider: p.driveProvider,
			IsCloudMode:   true,
		})
	} else {
		newInstance, err = NewStorage(userID, p.dataDir)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage for user %s: %w", userID, err)
	}

	instance = newInstance

	p.instances[userID] = instance
	return instance, nil
}

// CloseAll closes all storage instances
func (p *Provider) CloseAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for userID, instance := range p.instances {
		if err := instance.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close storage for user %s: %w", userID, err))
		}
		delete(p.instances, userID)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing storage instances: %v", errs)
	}

	return nil
}

// CleanupInactive removes storage instances that haven't been accessed recently
func (p *Provider) CleanupInactive(inactiveDuration time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-inactiveDuration)
	var toRemove []string

	for userID, instance := range p.instances {
		if instance.GetLastSyncTime().Before(cutoff) {
			toRemove = append(toRemove, userID)
		}
	}

	var errs []error
	for _, userID := range toRemove {
		if err := p.instances[userID].Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close storage for user %s: %w", userID, err))
		}
		delete(p.instances, userID)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}

	return nil
}

// GetActiveUsers returns the list of users with active storage instances
func (p *Provider) GetActiveUsers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	users := make([]string, 0, len(p.instances))
	for userID := range p.instances {
		users = append(users, userID)
	}
	return users
}
