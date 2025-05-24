package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benidevo/ascentio/internal/auth/models"
	authrepo "github.com/benidevo/ascentio/internal/auth/repository"
	authservice "github.com/benidevo/ascentio/internal/auth/services"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

// UserConfig represents the user configuration structure in the config file
type UserConfig struct {
	Users []UserEntry `yaml:"users" json:"users"`
}

// UserEntry represents a single user entry in the config file
type UserEntry struct {
	Username       string `yaml:"username" json:"username"`
	Password       string `yaml:"password" json:"password"`
	Email          string `yaml:"email" json:"email"`
	Role           string `yaml:"role" json:"role"`
	ResetOnNextRun bool   `yaml:"reset_on_next_run" json:"reset_on_next_run"`
}

// UserSyncCommand handles user synchronization from config files
type UserSyncCommand struct {
	db      *sql.DB
	config  *config.Settings
	log     zerolog.Logger
	authSvc *authservice.AuthService
}

// NewUserSyncCommand creates a new user sync command
func NewUserSyncCommand(db *sql.DB, cfg *config.Settings) *UserSyncCommand {
	return &UserSyncCommand{
		db:      db,
		config:  cfg,
		log:     logger.GetLogger("user-sync"),
		authSvc: authservice.NewAuthService(authrepo.NewSQLiteUserRepository(db), cfg),
	}
}

// Name returns the command name
func (c *UserSyncCommand) Name() string {
	return "sync-users"
}

// Description returns the command description
func (c *UserSyncCommand) Description() string {
	return "Synchronize users from config file"
}

// Execute runs the user sync command
func (c *UserSyncCommand) Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config file path required")
	}

	configPath := args[0]
	c.log.Info().Str("path", configPath).Msg("Loading user config file")

	userConfig, err := c.loadUserConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	if len(userConfig.Users) == 0 {
		c.log.Warn().Msg("No users found in config file")
		return nil
	}

	for _, user := range userConfig.Users {
		if err := c.processUser(user); err != nil {
			c.log.Error().Err(err).Str("username", user.Username).Msg("Failed to process user")
			continue
		}
	}

	return nil
}

func (c *UserSyncCommand) loadUserConfig(path string) (*UserConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config UserConfig
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(file, &config); err != nil {
			return nil, err
		}
	case ".json":
		if err := json.Unmarshal(file, &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return &config, nil
}

func (c *UserSyncCommand) processUser(entry UserEntry) error {
	ctx := context.Background()
	role := entry.Role
	if role == "" {
		role = "standard"
	}

	user, err := c.authSvc.Register(ctx, entry.Username, entry.Password, role)
	if err != nil && err != models.ErrUserAlreadyExists {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err == nil {
		c.log.Info().Str("username", entry.Username).Msg("User created successfully")
		return nil
	}

	if entry.ResetOnNextRun {
		c.log.Info().Str("username", entry.Username).Msg("Resetting user password")

		if err := c.authSvc.ChangePassword(ctx, user.ID, entry.Password); err != nil {
			return fmt.Errorf("failed to reset password: %w", err)
		}

		c.log.Info().Str("username", entry.Username).Msg("Password reset successfully")
	} else {
		c.log.Info().Str("username", entry.Username).Msg("User exists, no action required")
	}

	return nil
}
