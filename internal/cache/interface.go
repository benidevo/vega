package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string, value any) error

	// Set stores a value in cache with optional TTL
	// ttl = 0 means no expiration
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete removes one or more keys from cache
	Delete(ctx context.Context, keys ...string) error

	// DeletePattern removes all keys matching the pattern
	// Pattern supports wildcards: "user:*" matches all keys starting with "user:"
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close gracefully shuts down the cache
	Close() error
}
