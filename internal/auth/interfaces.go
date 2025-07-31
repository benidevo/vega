package auth

import (
	"context"

	"github.com/benidevo/vega/internal/auth/models"
	"github.com/benidevo/vega/internal/auth/services"
)

// LoginService handles user authentication operations
type LoginService interface {
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	LogError(err error)
}

// TokenService handles token-related operations
type TokenService interface {
	VerifyToken(token string) (*services.Claims, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
}

// UserService handles user management operations
type UserService interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	Register(ctx context.Context, username, password, role string) (*models.User, error)
	ChangePassword(ctx context.Context, userID int, newPassword string) error
}

// FullAuthService combines all auth-related interfaces
// This is useful when a component needs multiple auth capabilities
type FullAuthService interface {
	LoginService
	TokenService
	UserService
}

// OAuthService handles OAuth authentication operations
type OAuthService interface {
	Authenticate(ctx context.Context, code, redirectURI string) (accessToken, refreshToken string, err error)
}
