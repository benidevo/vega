package scheduler

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/benidevo/prospector/internal/config"
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

func TestRegisterCommand(t *testing.T) {
	app, mockCmd := setupApp(t)

	assert.Len(t, app.commands, 1)
	assert.Contains(t, app.commands, "test-command")
	assert.Equal(t, mockCmd, app.commands["test-command"])
}

func TestRunCommand(t *testing.T) {
	t.Run("successfully executes a command", func(t *testing.T) {
		app, mockCmd := setupApp(t)
		defer app.db.Close()

		cmdArgs := []string{"arg1", "arg2"}
		mockCmd.On("Execute", cmdArgs).Return(nil).Once()

		err := app.Run([]string{"test-command", "arg1", "arg2"})

		assert.NoError(t, err)
		mockCmd.AssertExpectations(t)
	})

	t.Run("returns error for unknown command", func(t *testing.T) {
		app, _ := setupApp(t)
		defer app.db.Close()

		err := app.Run([]string{"unknown-command"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("returns error when command execution fails", func(t *testing.T) {
		app, mockCmd := setupApp(t)
		defer app.db.Close()

		expectedErr := errors.New("command execution failed")
		mockCmd.On("Execute", []string{}).Return(expectedErr).Once()

		err := app.Run([]string{"test-command"})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockCmd.AssertExpectations(t)
	})
}

func TestShowHelp(t *testing.T) {
	app, _ := setupApp(t)
	defer app.db.Close()

	// Execute the app with no arguments to show help
	err := app.Run([]string{})

	assert.NoError(t, err)
}

func TestNewApp(t *testing.T) {
	cfg := &config.Settings{}
	app := NewApp(cfg)

	assert.NotNil(t, app)
	assert.Equal(t, cfg, app.config)
	assert.NotNil(t, app.commands)
	assert.Len(t, app.commands, 0)
}

func TestSetupDB(t *testing.T) {
	cfg := &config.Settings{
		DBDriver:           "sqlite",
		DBConnectionString: ":memory:",
	}

	app := NewApp(cfg)

	err := app.setupDB()
	assert.NoError(t, err)
	assert.NotNil(t, app.db)

	app.db.Close()
}

// This is a simple test to ensure registerCommands doesn't panic
func TestRegisterCommands(t *testing.T) {
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
}
