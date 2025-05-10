package commands

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/benidevo/prospector/internal/auth"
	"github.com/benidevo/prospector/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// AuthServiceInterface captures the methods from auth.AuthService needed for testing
type AuthServiceInterface interface {
	Register(ctx context.Context, username, password, role string) (*auth.User, error)
	ChangePassword(ctx context.Context, userID int, newPassword string) error
}

// MockAuthService mocks the AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, username, password, role string) (*auth.User, error) {
	args := m.Called(ctx, username, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID int) (*auth.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthService) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthService) GenerateToken(user *auth.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) VerifyToken(token string) (*auth.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	args := m.Called(ctx, userID, newPassword)
	return args.Error(0)
}

// TestUserSyncCommand is a modified version for testing that accepts an interface
type TestUserSyncCommand struct {
	*UserSyncCommand
	mockAuthSvc AuthServiceInterface
}

// Process user overrides the original method to use the mock
func (c *TestUserSyncCommand) processUser(entry UserEntry) error {
	ctx := context.Background()
	role := entry.Role
	if role == "" {
		role = "standard"
	}

	user, err := c.mockAuthSvc.Register(ctx, entry.Username, entry.Password, role)
	if err != nil && err != auth.ErrUserAlreadyExists {
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

func createTestLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Str("module", "user-sync-test").Logger()
}

func setupTest(t *testing.T) (*MockAuthService, *TestUserSyncCommand, string) {
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

	cmd := &TestUserSyncCommand{
		UserSyncCommand: realCmd,
		mockAuthSvc:     mockService,
	}

	return mockService, cmd, configPath
}

func TestLoadUserConfig(t *testing.T) {
	_, command, configPath := setupTest(t)

	t.Run("loads valid YAML config", func(t *testing.T) {
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

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := command.loadUserConfig("/path/does/not/exist.yaml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		invalidPath := filepath.Join(t.TempDir(), "invalid.txt")
		err := os.WriteFile(invalidPath, []byte("not a yaml or json file"), 0644)
		require.NoError(t, err)

		_, err = command.loadUserConfig(invalidPath)
		assert.Error(t, err)
	})
}

func TestProcessUser(t *testing.T) {
	mockService, command, _ := setupTest(t)
	ctx := context.Background()

	t.Run("creates new user successfully", func(t *testing.T) {
		mockService.On("Register", ctx, "newuser", "password123", "standard").Return(&auth.User{
			ID:       1,
			Username: "newuser",
			Role:     auth.STANDARD,
		}, nil).Once()

		entry := UserEntry{
			Username: "newuser",
			Password: "password123",
			Email:    "new@example.com",
			Role:     "standard",
		}

		err := command.processUser(entry)

		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("handles existing user with password reset", func(t *testing.T) {
		existingUser := &auth.User{
			ID:       3,
			Username: "resetuser",
			Role:     auth.STANDARD,
		}

		mockService.On("Register", ctx, "resetuser", "oldpassword", "standard").Return(
			existingUser, auth.ErrUserAlreadyExists).Once()
		mockService.On("ChangePassword", ctx, 3, "oldpassword").Return(nil).Once()

		entry := UserEntry{
			Username:       "resetuser",
			Password:       "oldpassword",
			Role:           "standard",
			ResetOnNextRun: true,
		}

		err := command.processUser(entry)

		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("returns error when register fails", func(t *testing.T) {
		mockService.On("Register", ctx, "failuser", "password", "standard").Return(
			nil, auth.ErrUserCreationFailed).Once()

		entry := UserEntry{
			Username: "failuser",
			Password: "password",
			Role:     "standard",
		}

		err := command.processUser(entry)

		assert.Error(t, err)
		mockService.AssertExpectations(t)
	})
}

// Execute is skipped in TestUserSyncCommand tests since it would call the real processUser
func TestNewUserSyncCommand(t *testing.T) {
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
}
