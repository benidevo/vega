package home

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/vega/internal/cache"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Settings
		cache      cache.Cache
		jobService *job.JobService
	}{
		{
			name: "should_initialize_home_handler_with_dependencies",
			cfg: &config.Settings{
				IsCloudMode: false,
			},
			cache:      cache.NewNoOpCache(),
			jobService: &job.JobService{},
		},
		{
			name: "should_initialize_home_handler_in_cloud_mode",
			cfg: &config.Settings{
				IsCloudMode:        true,
				GoogleOAuthEnabled: true,
			},
			cache:      cache.NewNoOpCache(),
			jobService: &job.JobService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			handler := Setup(db, tt.cfg, tt.cache, tt.jobService)

			assert.NotNil(t, handler)
			assert.NotNil(t, handler.service)
			assert.Equal(t, tt.cfg, handler.cfg)
			assert.NotNil(t, handler.renderer)
		})
	}
}

func TestSetupService(t *testing.T) {
	tests := []struct {
		name       string
		cache      cache.Cache
		jobService *job.JobService
	}{
		{
			name:       "should_initialize_home_service_with_repositories",
			cache:      cache.NewNoOpCache(),
			jobService: &job.JobService{},
		},
		{
			name:       "should_initialize_home_service_without_cache",
			cache:      nil,
			jobService: &job.JobService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			service := SetupService(db, tt.cache, tt.jobService)

			assert.NotNil(t, service)
			assert.NotNil(t, service.jobRepository)
			assert.NotNil(t, service.jobService)
		})
	}
}

func TestSetup_Integration(t *testing.T) {
	t.Run("should_create_fully_functional_home_handler", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		cfg := &config.Settings{
			IsCloudMode:        false,
			GoogleOAuthEnabled: false,
		}

		cache := cache.NewNoOpCache()

		jobService := &job.JobService{}

		handler := Setup(db, cfg, cache, jobService)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.service)
		assert.NotNil(t, handler.service.jobRepository)
		assert.NotNil(t, handler.service.jobService)
		assert.Equal(t, cfg, handler.cfg)
		assert.NotNil(t, handler.renderer)
	})
}
