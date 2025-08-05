package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func setupTestCache(t *testing.T) (*BadgerCache, func()) {
	tempDir, err := os.MkdirTemp("", "cache_test_*")
	require.NoError(t, err)

	cache, err := NewBadgerCache(tempDir, 64)
	require.NoError(t, err)

	cleanup := func() {
		cache.Close()
		os.RemoveAll(tempDir)
	}

	return cache, cleanup
}

func TestBadgerCache_GetSet(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	key := "test:key"
	value := &testStruct{
		ID:   1,
		Name: "Test Item",
		Tags: []string{"tag1", "tag2"},
	}

	err := cache.Set(ctx, key, value, 0)
	assert.NoError(t, err)

	var retrieved testStruct
	err = cache.Get(ctx, key, &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, value.ID, retrieved.ID)
	assert.Equal(t, value.Name, retrieved.Name)
	assert.Equal(t, value.Tags, retrieved.Tags)
}

func TestBadgerCache_CacheMiss(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	var value testStruct
	err := cache.Get(ctx, "non:existent", &value)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestBadgerCache_SetWithTTL(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	testCases := []struct {
		name string
		ttl  time.Duration
	}{
		{"with TTL", 1 * time.Hour},
		{"without TTL", 0},
		{"short TTL", 1 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := "ttl:" + tc.name
			value := &testStruct{ID: 1, Name: tc.name}

			err := cache.Set(ctx, key, value, tc.ttl)
			assert.NoError(t, err)

			var retrieved testStruct
			err = cache.Get(ctx, key, &retrieved)
			assert.NoError(t, err)
			assert.Equal(t, value.ID, retrieved.ID)
		})
	}
}

func TestBadgerCache_Delete(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	keys := []string{"del:1", "del:2", "del:3"}
	for i, key := range keys {
		err := cache.Set(ctx, key, &testStruct{ID: i}, 0)
		assert.NoError(t, err)
	}

	err := cache.Delete(ctx, keys[0], keys[1])
	assert.NoError(t, err)

	var value testStruct
	assert.Equal(t, ErrCacheMiss, cache.Get(ctx, keys[0], &value))
	assert.Equal(t, ErrCacheMiss, cache.Get(ctx, keys[1], &value))

	err = cache.Get(ctx, keys[2], &value)
	assert.NoError(t, err)
	assert.Equal(t, 2, value.ID)
}

func TestBadgerCache_DeletePattern(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	testData := map[string]int{
		"user:1:profile": 1,
		"user:1:stats":   2,
		"user:2:profile": 3,
		"user:2:stats":   4,
		"job:1:details":  5,
	}

	for key, id := range testData {
		err := cache.Set(ctx, key, &testStruct{ID: id}, 0)
		assert.NoError(t, err)
	}

	err := cache.DeletePattern(ctx, "user:1:*")
	assert.NoError(t, err)

	var value testStruct

	assert.Equal(t, ErrCacheMiss, cache.Get(ctx, "user:1:profile", &value))
	assert.Equal(t, ErrCacheMiss, cache.Get(ctx, "user:1:stats", &value))

	assert.NoError(t, cache.Get(ctx, "user:2:profile", &value))
	assert.Equal(t, 3, value.ID)

	assert.NoError(t, cache.Get(ctx, "user:2:stats", &value))
	assert.Equal(t, 4, value.ID)

	assert.NoError(t, cache.Get(ctx, "job:1:details", &value))
	assert.Equal(t, 5, value.ID)
}

func TestBadgerCache_Exists(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	key := "exists:test"

	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	err = cache.Set(ctx, key, &testStruct{ID: 1}, 0)
	assert.NoError(t, err)

	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)

	err = cache.Delete(ctx, key)
	assert.NoError(t, err)

	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)
}
