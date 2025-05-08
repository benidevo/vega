package auth

import (
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo   UserRepository
	config *config.Settings
	log    zerolog.Logger
}

func NewAuthService(repo UserRepository, config *config.Settings) *AuthService {
	return &AuthService{repo: repo, config: config, log: logger.GetLogger("auth")}
}

// Register creates a new user with the provided username, password, and role.
func (s *AuthService) Register(username, password, role string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		return nil, ErrUserCreationFailed
	}

	user, err := s.repo.CreateUser(username, hashedPassword, role)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create user")
		return nil, ErrUserCreationFailed
	}

	return user, nil
}

// Login authenticates a user by verifying the provided username and password.
//
// If successful, it generates and returns a JWT token. It also updates the
// user's last login timestamp.
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		s.log.Error().Err(err).Msg("User not found")
		return "", ErrUserNotFound
	}

	if !verifyPassword(user.Password, password) {
		s.log.Error().Msg("Invalid credentials")
		return "", ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate token")
		return "", err
	}

	user.LastLogin = time.Now()
	_, err = s.repo.UpdateUser(user)
	if err != nil {
		s.log.Warn().Err(err).Msg("Failed to update user last login")
	}

	s.log.Info().Msgf("User %s logged in successfully", username)
	return token, nil
}

// GetUserByID retrieves a user by their unique ID from the repository.
func (s *AuthService) GetUserByID(userID int) (*User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to find user by ID")
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GenerateToken generates a JWT token for the given user with claims including
// user ID, username, role, and standard registered claims such as issuer, subject,
// issued time, and expiration time.
//
// The token is signed using the HS256 algorithm
// and a secret key from the service configuration.
func (s *AuthService) GenerateToken(user *User) (string, error) {
	expirationTime := time.Now().Add(s.config.TokenExpiration)
	role, _ := user.Role.String()

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "prospector",
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.TokenSecret))
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
func (s *AuthService) ChangePassword(userID int, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		return ErrUserPasswordChangeFailed
	}

	user.Password = hashedPassword
	_, err = s.repo.UpdateUser(user)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to update user password")
		return ErrUserPasswordChangeFailed
	}

	return nil
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
