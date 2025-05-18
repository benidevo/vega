package auth

import (
	"context"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
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
	repo   UserRepository
	config *config.Settings
	log    zerolog.Logger
}

// NewAuthService creates and returns a new AuthService instance using the provided UserRepository and configuration settings.
func NewAuthService(repo UserRepository, config *config.Settings) *AuthService {
	return &AuthService{repo: repo, config: config, log: logger.GetLogger("auth")}
}

// Register creates a new user with the provided username, password, and role.
// It returns the user and ErrUserAlreadyExists if a user with that username already exists.
func (s *AuthService) Register(ctx context.Context, username, password, role string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		return nil, ErrUserCreationFailed
	}

	user, err := s.repo.CreateUser(ctx, username, hashedPassword, role)
	if err != nil {
		sentinelErr := GetSentinelError(err)

		if sentinelErr == ErrUserAlreadyExists {
			s.log.Warn().Str("username", username).Msg("User already exists")
			return user, ErrUserAlreadyExists
		}

		s.log.Error().Err(err).Str("username", username).Msg("Failed to create user")
		return nil, sentinelErr
	}

	s.log.Info().Str("username", username).Msg("User registered successfully")
	return user, nil
}

// Login authenticates a user by verifying the provided username and password.
// If successful, it generates and returns access and refresh tokens and updates the user's last login timestamp.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		if sentinelErr == ErrUserNotFound {
			s.log.Info().Str("username", username).Msg("Login attempt for non-existent user")
		} else {
			s.log.Error().Err(err).Str("username", username).Msg("Error retrieving user during login")
		}

		return "", "", ErrInvalidCredentials
	}

	if user.Password == "" {
		s.log.Error().Str("username", username).Msg("User password is empty. Account was created using Google authentication")
		return "", "", ErrInvalidCredentials
	}

	if !verifyPassword(user.Password, password) {
		s.log.Error().Str("username", username).Msg("Invalid password provided during login")
		return "", "", ErrInvalidCredentials
	}

	accessToken, err := GenerateAccessToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).Str("username", username).Msg("Failed to generate access token")
		return "", "", ErrTokenCreationFailed
	}

	refreshToken, err := GenerateRefreshToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).Str("username", username).Msg("Failed to generate refresh token")
		return "", "", ErrTokenCreationFailed
	}

	user.LastLogin = time.Now().UTC()
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		s.log.Warn().Err(err).Str("username", username).Str("error_type", sentinelErr.Error()).Msg("Failed to update user last login")
	}

	s.log.Info().Str("username", username).Int("user_id", user.ID).Msg("User logged in successfully")
	return accessToken, refreshToken, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token if valid
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.VerifyToken(refreshToken)
	if err != nil {
		s.log.Error().Err(err).Msg("Invalid refresh token")
		return "", ErrInvalidToken
	}

	if claims.TokenType != "refresh" {
		s.log.Error().Msg("Token provided is not a refresh token")
		return "", ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		s.log.Error().Err(err).Int("user_id", claims.UserID).Str("error_type", sentinelErr.Error()).Msg("Failed to find user for token refresh")
		return "", ErrInvalidToken
	}

	accessToken, err := GenerateAccessToken(user, s.config)
	if err != nil {
		s.log.Error().Err(err).Int("user_id", user.ID).Msg("Failed to generate new access token")
		return "", ErrTokenCreationFailed
	}

	s.log.Info().Int("user_id", user.ID).Msg("Access token refreshed successfully")
	return accessToken, nil
}

// GetUserByID retrieves a user by their unique ID from the repository.
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		s.log.Error().Err(err).Int("user_id", userID).Str("error_type", sentinelErr.Error()).Msg("Failed to find user by ID")

		if sentinelErr == ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, ErrUserRetrievalFailed
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
			return nil, ErrInvalidToken
		}
		return []byte(s.config.TokenSecret), nil
	})

	if err != nil {
		s.log.Error().Err(err).Msg("Failed to parse token")
		return nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		s.log.Error().Msg("Token is not valid")
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ChangePassword updates the password for a user identified by userID.
//
// It hashes the new password and updates the user's record in the repository.
func (s *AuthService) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		s.log.Error().Err(err).Int("user_id", userID).Str("error_type", sentinelErr.Error()).Msg("Failed to find user for password change")
		if sentinelErr == ErrUserNotFound {
			return ErrUserNotFound
		}
		return ErrUserPasswordChangeFailed
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		s.log.Error().Err(err).Int("user_id", userID).Msg("Failed to hash password")
		return ErrUserPasswordChangeFailed
	}

	user.Password = hashedPassword
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := GetSentinelError(err)
		s.log.Error().Err(err).Int("user_id", userID).Str("error_type", sentinelErr.Error()).Msg("Failed to update user password")
		return ErrUserPasswordChangeFailed
	}

	s.log.Info().Int("user_id", userID).Msg("User password changed successfully")
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
func GenerateToken(user *User, cfg *config.Settings, tokenType TokenType, expiry time.Duration) (string, error) {
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
func GenerateAccessToken(user *User, cfg *config.Settings) (string, error) {
	return GenerateToken(user, cfg, AccessToken, cfg.AccessTokenExpiry)
}

// GenerateRefreshToken creates a long-lived refresh JWT token for the given user.
func GenerateRefreshToken(user *User, cfg *config.Settings) (string, error) {
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
