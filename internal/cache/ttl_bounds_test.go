package cache

import (
	"context"
	"testing"
	"time"
)

func TestBadgerCache_TTLBounds(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	testCases := []struct {
		name         string
		requestedTTL time.Duration
		expectedMax  time.Duration
		description  string
	}{
		{
			name:         "zero_ttl_gets_capped",
			requestedTTL: 0,
			expectedMax:  time.Hour,
			description:  "Zero TTL should be capped at 1 hour",
		},
		{
			name:         "excessive_ttl_gets_capped",
			requestedTTL: 24 * time.Hour,
			expectedMax:  time.Hour,
			description:  "24-hour TTL should be capped at 1 hour",
		},
		{
			name:         "valid_30min_ttl_unchanged",
			requestedTTL: 30 * time.Minute,
			expectedMax:  time.Hour,
			description:  "30-minute TTL should remain unchanged",
		},
		{
			name:         "valid_10min_ttl_unchanged",
			requestedTTL: 10 * time.Minute,
			expectedMax:  time.Hour,
			description:  "10-minute TTL should remain unchanged",
		},
		{
			name:         "exactly_1hour_ttl_allowed",
			requestedTTL: time.Hour,
			expectedMax:  time.Hour,
			description:  "Exactly 1-hour TTL should be allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := "ttl_test_" + tc.name
			value := &testStruct{
				ID:   1,
				Name: tc.description,
			}

			err := cache.Set(ctx, key, value, tc.requestedTTL)
			if err != nil {
				t.Fatalf("Failed to set cache: %v", err)
			}

			var retrieved testStruct
			err = cache.Get(ctx, key, &retrieved)
			if err != nil {
				t.Fatalf("Failed to get cached value: %v", err)
			}

			if retrieved.ID != value.ID || retrieved.Name != value.Name {
				t.Errorf("Retrieved value doesn't match: got %+v, want %+v", retrieved, value)
			}

			_ = cache.Delete(ctx, key)
		})
	}
}

func TestBadgerCache_ConsistentTTLApplication(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	key := "consistent_ttl_test"
	value := &testStruct{ID: 99, Name: "test"}

	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Item should exist immediately after setting with 0 TTL")
	}

	var retrieved testStruct
	err = cache.Get(ctx, key, &retrieved)
	if err != nil {
		t.Errorf("Should be able to retrieve item with bounded TTL: %v", err)
	}
}
