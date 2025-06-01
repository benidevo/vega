package services

import (
	"context"
	"fmt"
	"time"

	"github.com/benidevo/ascentio/internal/auth/models"
	"github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents the JWT claims for authentication, including user ID, username, role, and standard registered claims.
type Claims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type,omitempty"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthService provides authentication and user management functionality
type AuthService struct {
	repo   repository.UserRepository
	config *config.Settings
	log    *logger.PrivacyLogger
}

// NewAuthService creates and returns a new AuthService instance using the provided UserRepository and configuration settings.
func NewAuthService(repo repository.UserRepository, config *config.Settings) *AuthService {
	return &AuthService{repo: repo, config: config, log: logger.GetPrivacyLogger("auth")}
}

// LogError logs an authentication error using the service's logger.
func (s *AuthService) LogError(err error) {
	s.log.Error().Err(err).Msg("Authentication error")
}

// Register creates a new user with the provided username, password, and role.
// It returns the user and models.ErrUserAlreadyExists if a user with that username already exists.
func (s *AuthService) Register(ctx context.Context, username, password, role string) (*models.User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		return nil, models.ErrUserCreationFailed
	}

	user, err := s.repo.CreateUser(ctx, username, hashedPassword, role)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)

		if sentinelErr == models.ErrUserAlreadyExists {
			s.log.Warn().
				Str("event", "user_registration_duplicate").
				Str("hashed_id", logger.HashIdentifier(username)).
				Msg("User already exists")
			return user, models.ErrUserAlreadyExists
		}

		s.log.Error().Err(err).
			Str("event", "user_registration_failed").
			Str("hashed_id", logger.HashIdentifier(username)).
			Msg("Failed to create user")
		return nil, sentinelErr
	}

	s.log.LogRegistrationEvent("user_registered", logger.HashIdentifier(username), true)
	return user, nil
}

// Login authenticates a user by verifying the provided username and password.
// If successful, it generates and returns access and refresh tokens and updates the user's last login timestamp.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		if sentinelErr == models.ErrUserNotFound {
			s.log.Info().
				Str("event", "login_attempt_unknown_user").
				Str("hashed_id", logger.HashIdentifier(username)).
				Msg("Login attempt for non-existent user")
		} else {
			s.log.Error().Err(err).
				Str("event", "login_user_retrieval_error").
				Str("hashed_id", logger.HashIdentifier(username)).
				Msg("Error retrieving user during login")
		}

		return "", "", models.ErrInvalidCredentials
	}

	if user.Password == "" {
		s.log.Error().
			Str("event", "login_oauth_account_password_attempt").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("User password is empty. Account was created using Google authentication")
		return "", "", models.ErrInvalidCredentials
	}

	if !verifyPassword(user.Password, password) {
		s.log.LogAuthEvent("login_invalid_password", user.ID, false)
		return "", "", models.ErrInvalidCredentials
	}

	accessToken, err := GenerateAccessToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "access_token_generation_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("Failed to generate access token")
		return "", "", models.ErrInvalidCredentials
	}

	refreshToken, err := GenerateRefreshToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "refresh_token_generation_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("Failed to generate refresh token")
		return "", "", models.ErrInvalidCredentials
	}

	user.LastLogin = time.Now().UTC()
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Warn().Err(err).
			Str("event", "last_login_update_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to update user last login")
	}

	s.log.LogAuthEvent("login_success", user.ID, true)
	return accessToken, refreshToken, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token if valid
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.VerifyToken(refreshToken)
	if err != nil {
		s.log.Error().Err(err).Msg("Invalid refresh token")
		return "", models.ErrInvalidToken
	}

	if claims.TokenType != "refresh" {
		s.log.Error().Msg("Token provided is not a refresh token")
		return "", models.ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Error().Err(err).
			Str("event", "token_refresh_user_not_found").
			Str("user_ref", fmt.Sprintf("user_%d", claims.UserID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to find user for token refresh")
		return "", models.ErrInvalidToken
	}

	accessToken, err := GenerateAccessToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "token_refresh_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("Failed to generate new access token")
		return "", models.ErrTokenCreationFailed
	}

	s.log.Info().
		Str("event", "token_refreshed").
		Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
		Msg("Access token refreshed successfully")
	return accessToken, nil
}

// GetUserByID retrieves a user by their unique ID from the repository.
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Error().Err(err).
			Str("event", "user_lookup_failed").
			Str("user_ref", fmt.Sprintf("user_%d", userID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to find user by ID")

		if sentinelErr == models.ErrUserNotFound {
			return nil, models.ErrUserNotFound
		}
		return nil, models.ErrUserRetrievalFailed
	}
	return user, nil
}

// VerifyToken validates a JWT token and extracts its claims.
//
// It checks the token's signing method, parses it with the provided claims,
// and ensures the token is valid.
func (s *AuthService) VerifyToken(token string) (*Claims, error) {
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if jwtToken.Method != jwt.SigningMethodHS256 {
			s.log.Error().Msg("Unexpected signing method")
			return nil, models.ErrInvalidToken
		}
		return []byte(s.config.TokenSecret), nil
	})

	if err != nil {
		s.log.Error().Err(err).Msg("Failed to parse token")
		return nil, models.ErrInvalidToken
	}

	if !parsedToken.Valid {
		s.log.Error().Msg("Token is not valid")
		return nil, models.ErrInvalidToken
	}

	return claims, nil
}

// ChangePassword updates the password for a user identified by userID.
//
// It hashes the new password and updates the user's record in the repository.
func (s *AuthService) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Error().Err(err).
			Str("event", "password_change_user_not_found").
			Str("user_ref", fmt.Sprintf("user_%d", userID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to find user for password change")
		if sentinelErr == models.ErrUserNotFound {
			return models.ErrUserNotFound
		}
		return models.ErrUserPasswordChangeFailed
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "password_hash_failed").
			Str("user_ref", fmt.Sprintf("user_%d", userID)).
			Msg("Failed to hash password")
		return models.ErrUserPasswordChangeFailed
	}

	user.Password = hashedPassword
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Error().Err(err).
			Str("event", "password_update_failed").
			Str("user_ref", fmt.Sprintf("user_%d", userID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to update user password")
		return models.ErrUserPasswordChangeFailed
	}

	s.log.Info().
		Str("event", "password_changed").
		Str("user_ref", fmt.Sprintf("user_%d", userID)).
		Msg("User password changed successfully")
	return nil
}

// TokenType defines the type of JWT token
type TokenType string

const (
	// AccessToken is a short-lived token used for authentication
	AccessToken TokenType = "access"
	// RefreshToken is a long-lived token used to refresh access tokens
	RefreshToken TokenType = "refresh"
)

// GenerateToken creates a JWT token for the given user with the specified token type and expiration duration.
func GenerateToken(user *models.User, cfg *config.Settings, tokenType TokenType, expiry time.Duration) (string, error) {
	expirationTime := time.Now().UTC().Add(expiry)
	role := user.Role.String()

	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      role,
		TokenType: string(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.AppName,
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.TokenSecret))
}

// GenerateAccessToken creates a short-lived access JWT token for the given user.
func GenerateAccessToken(user *models.User, cfg *config.Settings) (string, error) {
	return GenerateToken(user, cfg, AccessToken, cfg.AccessTokenExpiry)
}

// GenerateRefreshToken creates a long-lived refresh JWT token for the given user.
func GenerateRefreshToken(user *models.User, cfg *config.Settings) (string, error) {
	return GenerateToken(user, cfg, RefreshToken, cfg.RefreshTokenExpiry)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
