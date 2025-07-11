package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestNewFactory(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Settings
		wantErr bool
	}{
		{
			name: "successful creation with default config",
			config: &config.Settings{
				IsCloudMode: false,
			},
			wantErr: false,
		},
		{
			name: "successful creation with cloud mode",
			config: &config.Settings{
				IsCloudMode: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			factory, err := NewFactory(tt.config, db)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, factory)
			assert.NotNil(t, factory.GetProvider())
		})
	}
}

func TestFactory_GetUserStorage(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Settings{
		IsCloudMode: false,
	}

	factory, err := NewFactory(cfg, db)
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "valid user ID",
			userID:  "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := factory.GetUserStorage(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
			}
		})
	}
}

func TestFactory_Close(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Settings{
		IsCloudMode: false,
	}

	factory, err := NewFactory(cfg, db)
	require.NoError(t, err)

	// Should not error even when called multiple times
	err = factory.Close()
	assert.NoError(t, err)

	err = factory.Close()
	assert.NoError(t, err)
}

func TestTemporaryStorage_NotImplemented(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Settings{
		IsCloudMode: false,
	}

	factory, err := NewFactory(cfg, db)
	require.NoError(t, err)

	storage, err := factory.GetUserStorage(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, storage)

	// Test that all methods return "not implemented" error
	ctx := context.Background()

	_, err = storage.GetProfile(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	err = storage.SaveProfile(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	_, err = storage.ListCompanies(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	err = storage.Sync(ctx)
	assert.NoError(t, err)

	syncTime := storage.GetLastSyncTime()
	assert.NotZero(t, syncTime)

	err = storage.Initialize(ctx, "test@example.com")
	assert.NoError(t, err)

	err = storage.Close()
	assert.NoError(t, err)
}
