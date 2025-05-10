package scheduler

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/logger"
	"github.com/benidevo/prospector/internal/scheduler/commands"
	"github.com/rs/zerolog"
	_ "modernc.org/sqlite"
)

// Command represents a runnable scheduler command
type Command interface {
	Name() string
	Description() string
	Execute(args []string) error
}

// App represents the scheduler application
type App struct {
	config   *config.Settings
	commands map[string]Command
	log      zerolog.Logger
	db       *sql.DB
}

// NewApp creates a new scheduler app instance
func NewApp(cfg *config.Settings) *App {
	log := logger.GetLogger("scheduler")

	app := &App{
		config:   cfg,
		commands: make(map[string]Command),
		log:      log,
	}

	return app
}

func (a *App) registerCommands() {
	a.RegisterCommand(commands.NewUserSyncCommand(a.db, a.config))
}

// RegisterCommand adds a command to the scheduler
func (a *App) RegisterCommand(cmd Command) {
	a.commands[cmd.Name()] = cmd
}

func (a *App) setupDB() error {
	db, err := sql.Open(a.config.DBDriver, a.config.DBConnectionString)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	a.db = db
	return nil
}

// Run executes the scheduler with the provided arguments
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return a.showHelp()
	}

	if err := a.setupDB(); err != nil {
		return fmt.Errorf("failed to set up database: %w", err)
	}
	defer a.db.Close()

	a.registerCommands()

	cmdName := args[0]
	cmd, exists := a.commands[cmdName]

	if !exists {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	a.log.Info().Str("command", cmdName).Msg("Executing command")
	return cmd.Execute(args[1:])
}

func (a *App) showHelp() error {
	if err := a.setupDB(); err != nil {
		return fmt.Errorf("failed to set up database: %w", err)
	}
	defer a.db.Close()

	a.registerCommands()

	if len(a.commands) == 0 {
		return errors.New("no commands registered")
	}

	var b strings.Builder
	b.WriteString("Available commands:\n\n")

	for name, cmd := range a.commands {
		fmt.Fprintf(&b, "  %s\t%s\n", name, cmd.Description())
	}

	fmt.Println(b.String())
	return nil
}
