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

// SetupAuthWithService initializes and returns both AuthHandler and AuthService
// This allows other services to inject the auth service as a dependency
func SetupAuthWithService(db *sql.DB, cfg *config.Settings) (*AuthHandler, *services.AuthService) {
	repo := repository.NewSQLiteUserRepository(db)
	service := services.NewAuthService(repo, cfg)
	handler := NewAuthHandler(service, cfg)

	return handler, service
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

	// Check if using default credentials
	usingDefaults := cfg.AdminUsername == "admin" && cfg.AdminPassword == "VegaAdmin"

	user, err := authService.Register(ctx, cfg.AdminUsername, cfg.AdminPassword, "admin")
	if err != nil {
		log.Error().Err(err).
			Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
			Msg("Failed to create admin user")
		return err
	}

	log.Info().
		Str("hashed_id", logger.HashIdentifier(cfg.AdminUsername)).
		Int("user_id", user.ID).
		Msg("Admin user created successfully")

	if usingDefaults {
		log.Warn().
			Msg("⚠️  DEFAULT CREDENTIALS IN USE: Username 'admin', Password 'VegaAdmin'")
		log.Warn().
			Msg("⚠️  Please change the default password immediately at /settings/security")
		log.Warn().
			Msg("⚠️  Or set ADMIN_USERNAME and ADMIN_PASSWORD environment variables")
	}

	return nil
}
