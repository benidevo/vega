package commands

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/benidevo/ascentio/internal/auth/models"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// AuthServiceInterface captures the methods from models.AuthService needed for testing
type AuthServiceInterface interface {
	Register(ctx context.Context, username, password, role string) (*models.User, error)
	ChangePassword(ctx context.Context, userID int, newPassword string) error
}

// MockAuthService mocks the AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, username, password, role string) (*models.User, error) {
	args := m.Called(ctx, username, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	args := m.Called(ctx, userID, newPassword)
	return args.Error(0)
}

// TestUserSyncCommand is a modified version for testing that accepts an interface
type TestUserSyncCommandWrapper struct {
	*UserSyncCommand
	mockAuthSvc AuthServiceInterface
}

// Process user overrides the original method to use the mock
func (c *TestUserSyncCommandWrapper) processUser(entry UserEntry) error {
	ctx := context.Background()
	role := entry.Role
	if role == "" {
		role = "standard"
	}

	user, err := c.mockAuthSvc.Register(ctx, entry.Username, entry.Password, role)
	if err != nil && err != models.ErrUserAlreadyExists {
		return err
	}

	if err == nil {
		c.log.Info().Str("username", entry.Username).Msg("User created successfully")
		return nil
	}

	if entry.ResetOnNextRun {
		c.log.Info().Str("username", entry.Username).Msg("Resetting user password")

		if err := c.mockAuthSvc.ChangePassword(ctx, user.ID, entry.Password); err != nil {
			return err
		}

		c.log.Info().Str("username", entry.Username).Msg("Password reset successfully")
	} else {
		c.log.Info().Str("username", entry.Username).Msg("User exists, no action required")
	}

	return nil
}

func createTestLogger() *logger.PrivacyLogger {
	return logger.GetPrivacyLogger("user-sync-test")
}

func setupTest(t *testing.T) (*MockAuthService, *TestUserSyncCommandWrapper, string) {
	mockService := new(MockAuthService)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "users.yaml")

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	realCmd := &UserSyncCommand{
		db:     db,
		config: &config.Settings{},
		log:    createTestLogger(),
	}

	cmd := &TestUserSyncCommandWrapper{
		UserSyncCommand: realCmd,
		mockAuthSvc:     mockService,
	}

	return mockService, cmd, configPath
}

func TestUserSyncOperations(t *testing.T) {
	t.Run("should_load_users_config_when_yaml_is_valid", func(t *testing.T) {
		_, command, configPath := setupTest(t)

		yamlContent := `users:
  - username: admin
    password: securepassword
    email: admin@example.com
    role: admin
    reset_on_next_run: false
  - username: user
    password: userpass
    email: user@example.com
    role: standard
    reset_on_next_run: true
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		config, err := command.loadUserConfig(configPath)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config.Users, 2)
		assert.Equal(t, "admin", config.Users[0].Username)
		assert.Equal(t, "securepassword", config.Users[0].Password)
		assert.Equal(t, "admin", config.Users[0].Role)
		assert.False(t, config.Users[0].ResetOnNextRun)
	})

	t.Run("should_return_error_when_config_file_not_found", func(t *testing.T) {
		_, command, _ := setupTest(t)

		_, err := command.loadUserConfig("/path/does/not/exist.yaml")
		assert.Error(t, err)
	})

	t.Run("should_return_error_when_config_file_format_invalid", func(t *testing.T) {
		_, command, _ := setupTest(t)

		invalidPath := filepath.Join(t.TempDir(), "invalid.txt")
		err := os.WriteFile(invalidPath, []byte("not a yaml or json file"), 0644)
		require.NoError(t, err)

		_, err2 := command.loadUserConfig(invalidPath)
		assert.Error(t, err2)
	})

	t.Run("should_create_new_user_when_user_does_not_exist", func(t *testing.T) {
		mockService, command, _ := setupTest(t)
		ctx := context.Background()

		mockService.On("Register", ctx, "newuser", "password123", "standard").Return(&models.User{
			ID:       1,
			Username: "newuser",
			Role:     models.STANDARD,
		}, nil).Once()

		err := command.processUser(UserEntry{
			Username: "newuser",
			Password: "password123",
			Role:     "standard",
		})
		assert.NoError(t, err)

		mockService.AssertExpectations(t)
	})

	t.Run("should_reset_password_when_user_exists_and_reset_enabled", func(t *testing.T) {
		mockService, command, _ := setupTest(t)
		ctx := context.Background()

		existingUser := &models.User{
			ID:       3,
			Username: "resetuser",
			Role:     models.STANDARD,
		}
		mockService.On("Register", ctx, "resetuser", "oldpassword", "standard").Return(
			existingUser, models.ErrUserAlreadyExists).Once()
		mockService.On("ChangePassword", ctx, 3, "oldpassword").Return(nil).Once()

		err := command.processUser(UserEntry{
			Username:       "resetuser",
			Password:       "oldpassword",
			Role:           "standard",
			ResetOnNextRun: true,
		})
		assert.NoError(t, err)

		mockService.AssertExpectations(t)
	})

	t.Run("should_return_error_when_user_creation_fails", func(t *testing.T) {
		mockService, command, _ := setupTest(t)
		ctx := context.Background()

		mockService.On("Register", ctx, "failuser", "password", "standard").Return(
			nil, models.ErrUserCreationFailed).Once()

		err := command.processUser(UserEntry{
			Username: "failuser",
			Password: "password",
			Role:     "standard",
		})

		assert.Error(t, err)
		mockService.AssertExpectations(t)
	})
}

func TestNewUserSyncCommand(t *testing.T) {
	t.Run("should_initialize_command_with_required_services", func(t *testing.T) {
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL,
				password TEXT NOT NULL,
				role INTEGER NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				last_login TIMESTAMP
			)
		`)
		require.NoError(t, err)

		cfg := &config.Settings{}

		command := NewUserSyncCommand(db, cfg)

		assert.NotNil(t, command)
		assert.NotNil(t, command.authSvc, "AuthService should be initialized")
		assert.Equal(t, "sync-users", command.Name())
		assert.Equal(t, "Synchronize users from config file", command.Description())
	})
}
