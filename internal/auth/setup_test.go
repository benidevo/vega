package auth

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAuth(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Settings
	}{
		{
			name: "should_create_auth_handler_with_valid_config",
			cfg: &config.Settings{
				TokenSecret:        "test-secret-key-minimum-32-chars-long",
				DBConnectionString: "test.db",
			},
		},
		{
			name: "should_handle_short_token_secret",
			cfg: &config.Settings{
				TokenSecret:        "short",
				DBConnectionString: "test.db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			handler := SetupAuth(db, tt.cfg)

			assert.NotNil(t, handler)
			assert.NotNil(t, handler.service)
			assert.Equal(t, tt.cfg, handler.cfg)
		})
	}
}

func TestSetupAuthWithService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Settings{
		TokenSecret:        "test-secret",
		DBConnectionString: "test.db",
	}

	handler, service := SetupAuthWithService(db, cfg)

	assert.NotNil(t, handler)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, handler.cfg)
	assert.Equal(t, handler.service, service)
}

func TestSetupGoogleAuth(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Settings
		expectError bool
	}{
		{
			name: "should_initialize_google_auth_when_valid_config",
			cfg: &config.Settings{
				GoogleClientID:     "test-client-id",
				GoogleClientSecret: "test-client-secret",
				TokenSecret:        "test-secret",
			},
			expectError: false,
		},
		{
			name: "should_return_error_when_missing_client_config",
			cfg: &config.Settings{
				TokenSecret: "test-secret",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			handler, err := SetupGoogleAuth(tt.cfg, db)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				if handler != nil {
					assert.NotNil(t, handler.service)
					assert.Equal(t, tt.cfg, handler.cfg)
				}
			}
		})
	}
}
