package auth

import (
	"context"
	"database/sql"

	"github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/auth/services"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/config"
	"github.com/rs/zerolog/log"
)

// SetupAuth initializes and returns an AuthHandler using the provided database connection and configuration settings.
// It sets up the user repository, authentication service, and handler dependencies.
func SetupAuth(db *sql.DB, cfg *config.Settings) *AuthHandler {
	repo := repository.NewSQLiteUserRepository(db)
	service := services.NewAuthService(repo, cfg)
	handler := NewAuthHandler(service, cfg)

	return handler
}

// SetupGoogleAuth initializes and returns a GoogleAuthHandler using the provided configuration settings.
// It sets up the GoogleAuthService and handler dependencies.
func SetupGoogleAuth(cfg *config.Settings, db *sql.DB) (*GoogleAuthHandler, error) {
	repo := repository.NewSQLiteUserRepository(db)
	service, err := services.NewGoogleAuthService(cfg, repo)
	if err != nil {
		return nil, err
	}
	handler := NewGoogleAuthHandler(service, cfg)

	return handler, nil
}

// CreateAdminUserIfRequired creates an admin user if the configuration specifies to do so
// and the admin user doesn't already exist. This should be called during application startup.
func CreateAdminUserIfRequired(db *sql.DB, cfg *config.Settings) error {
	if !cfg.CreateAdminUser {
		return nil
	}

	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Warn().Msg("CREATE_ADMIN_USER is true but ADMIN_USERNAME or ADMIN_PASSWORD is not set")
		return nil
	}

	repo := repository.NewSQLiteUserRepository(db)
	authService := services.NewAuthService(repo, cfg)

	ctx := context.Background()

	admin, err := repo.FindByUsername(ctx, cfg.AdminUsername)
	if err == nil {
		log.Info().
			Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
			Msg("Admin user already exists, skipping creation")

		if cfg.ResetAdminPassword {
			log.Info().
				Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
				Msg("Resetting admin user password")
			err = authService.ChangePassword(ctx, admin.ID, cfg.AdminPassword)
			if err != nil {
				log.Error().Err(err).
					Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
					Msg("Failed to reset admin user password")
				return err
			}
			log.Info().
				Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
				Msg("Admin user password reset successfully")
		}
		return nil
	}

	user, err := authService.Register(ctx, cfg.AdminUsername, cfg.AdminPassword, "admin")
	if err != nil {
		log.Error().Err(err).
			Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
			Msg("Failed to create admin user")
		return err
	}

	log.Info().
		Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
		Str("hashed_email", logger.HashIdentifier(cfg.AdminEmail)).
		Int("user_id", user.ID).
		Msg("Admin user created successfully")

	return nil
}
