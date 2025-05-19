package scheduler

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// MockCommand is a mock implementation of the Command interface for testing
type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCommand) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCommand) Execute(args []string) error {
	callArgs := m.Called(args)
	return callArgs.Error(0)
}

func setupApp(t *testing.T) (*App, *MockCommand) {
	cfg := &config.Settings{
		DBDriver:           "sqlite",
		DBConnectionString: ":memory:",
	}

	mockCmd := new(MockCommand)
	mockCmd.On("Name").Return("test-command").Maybe()
	mockCmd.On("Description").Return("Test command description").Maybe()

	// Create the app
	app := NewApp(cfg)

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	app.db = db

	app.RegisterCommand(mockCmd)

	return app, mockCmd
}

func TestApp(t *testing.T) {
	t.Run("should_register_command_when_valid", func(t *testing.T) {
		app, mockCmd := setupApp(t)
		defer app.db.Close()

		assert.Len(t, app.commands, 1)
		assert.Contains(t, app.commands, "test-command")
		assert.Equal(t, mockCmd, app.commands["test-command"])
	})

	t.Run("should_execute_command_when_name_exists", func(t *testing.T) {
		app, mockCmd := setupApp(t)
		defer app.db.Close()

		cmdArgs := []string{"arg1", "arg2"}
		mockCmd.On("Execute", cmdArgs).Return(nil).Once()

		err := app.Run([]string{"test-command", "arg1", "arg2"})

		assert.NoError(t, err)
		mockCmd.AssertExpectations(t)
	})

	t.Run("should_return_error_when_command_not_found", func(t *testing.T) {
		app, _ := setupApp(t)
		defer app.db.Close()

		err := app.Run([]string{"unknown-command"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("should_return_error_when_command_execution_fails", func(t *testing.T) {
		app, mockCmd := setupApp(t)
		defer app.db.Close()

		expectedErr := errors.New("command execution failed")
		mockCmd.On("Execute", []string{}).Return(expectedErr).Once()

		err := app.Run([]string{"test-command"})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockCmd.AssertExpectations(t)
	})

	t.Run("should_show_help_when_no_args_provided", func(t *testing.T) {
		app, _ := setupApp(t)
		defer app.db.Close()

		err := app.Run([]string{})
		assert.NoError(t, err)
	})

	t.Run("should_initialize_app_with_config", func(t *testing.T) {
		cfg := &config.Settings{
			DBDriver:           "sqlite",
			DBConnectionString: ":memory:",
		}

		app := NewApp(cfg)

		assert.NotNil(t, app)
		assert.Equal(t, cfg, app.config)
		assert.NotNil(t, app.commands)
		assert.Len(t, app.commands, 0)
	})

	t.Run("should_setup_database_when_valid_config", func(t *testing.T) {
		cfg := &config.Settings{
			DBDriver:           "sqlite",
			DBConnectionString: ":memory:",
		}

		app := NewApp(cfg)

		err := app.setupDB()
		assert.NoError(t, err)
		assert.NotNil(t, app.db)
		defer app.db.Close()
	})

	t.Run("should_register_commands_after_initialization", func(t *testing.T) {
		cfg := &config.Settings{
			DBDriver:           "sqlite",
			DBConnectionString: ":memory:",
		}

		app := NewApp(cfg)
		err := app.setupDB()
		require.NoError(t, err)
		defer app.db.Close()

		app.registerCommands()
		assert.NotEmpty(t, app.commands)

		_, hasUserSync := app.commands["sync-users"]
		assert.True(t, hasUserSync, "UserSyncCommand should be registered")
	})
}
