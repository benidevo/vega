package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	commonerrors "github.com/benidevo/prospector/internal/common/errors"
	"github.com/benidevo/prospector/internal/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, username, password, role string) (*User, error) {
	args := m.Called(ctx, username, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAllUsers(ctx context.Context) ([]*User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func setupTestConfig() *config.Settings {
	return &config.Settings{
		TokenSecret:     "test-secret-key",
		TokenExpiration: time.Hour,
	}
}

func TestRegisterUser(t *testing.T) {
	t.Run("should_register_user_successfully_when_valid_data", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()
		ctx := context.Background()

		mockRepo.On("CreateUser", ctx, "testuser", mock.AnythingOfType("string"), "admin").Return(&User{
			ID:       1,
			Username: "testuser",
			Role:     ADMIN,
		}, nil)

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.Register(ctx, "testuser", "password123", "admin")

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, "testuser", user.Username)
		require.Equal(t, ADMIN, user.Role)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should_return_error_when_repository_fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()
		ctx := context.Background()

		repoErr := commonerrors.WrapError(ErrUserCreationFailed, errors.New("repository error"))
		mockRepo.On("CreateUser", ctx, "testuser", mock.AnythingOfType("string"), "admin").Return(nil, repoErr)

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.Register(ctx, "testuser", "password123", "admin")

		require.Error(t, err)
		require.Equal(t, ErrUserCreationFailed, err)
		require.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	t.Run("should_login_successfully_when_credentials_valid", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()
		ctx := context.Background()

		hashedPassword, err := hashPassword("password123")
		require.NoError(t, err)

		mockRepo.On("FindByUsername", ctx, "testuser").Return(&User{
			ID:       1,
			Username: "testuser",
			Password: hashedPassword,
			Role:     ADMIN,
		}, nil)

		mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*auth.User")).Return(&User{}, nil)

		authService := NewAuthService(mockRepo, cfg)

		token, err := authService.Login(ctx, "testuser", "password123")

		require.NoError(t, err)
		require.NotEmpty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should_return_error_when_credentials_invalid", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()
		ctx := context.Background()

		mockRepo.On("FindByUsername", ctx, "nonexistent").Return(nil, ErrUserNotFound)

		hashedPassword, err := hashPassword("correctpassword")
		require.NoError(t, err)
		mockRepo.On("FindByUsername", ctx, "testuser").Return(&User{
			ID:       1,
			Username: "testuser",
			Password: hashedPassword,
			Role:     ADMIN,
		}, nil)

		authService := NewAuthService(mockRepo, cfg)

		token1, err1 := authService.Login(ctx, "nonexistent", "password123")
		require.Error(t, err1)
		require.Equal(t, ErrInvalidCredentials, err1)
		require.Empty(t, token1)

		token2, err2 := authService.Login(ctx, "testuser", "wrongpassword")
		require.Error(t, err2)
		require.Equal(t, ErrInvalidCredentials, err2)
		require.Empty(t, token2)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := setupTestConfig()
	ctx := context.Background()

	mockRepo.On("FindByID", ctx, 1).Return(&User{
		ID:       1,
		Username: "testuser",
		Role:     ADMIN,
	}, nil)

	mockRepo.On("FindByID", ctx, 999).Return(nil, ErrUserNotFound)

	authService := NewAuthService(mockRepo, cfg)

	t.Run("should_return_user_when_found", func(t *testing.T) {
		user, err := authService.GetUserByID(ctx, 1)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, 1, user.ID)
		require.Equal(t, "testuser", user.Username)
	})

	t.Run("should_return_error_when_user_not_found", func(t *testing.T) {
		user, err := authService.GetUserByID(ctx, 999)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, user)
	})

	mockRepo.AssertExpectations(t)
}

func TestVerifyToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := setupTestConfig()
	authService := NewAuthService(mockRepo, cfg)

	user := &User{
		ID:       1,
		Username: "testuser",
		Role:     ADMIN,
	}

	token, err := GenerateToken(user, cfg)
	require.NoError(t, err)

	t.Run("should_verify_token_when_valid", func(t *testing.T) {
		claims, err := authService.VerifyToken(token)

		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, 1, claims.UserID)
		require.Equal(t, "testuser", claims.Username)
	})

	t.Run("should_reject_token_when_invalid", func(t *testing.T) {
		claims, err := authService.VerifyToken("invalid.token.here")

		require.Error(t, err)
		require.Equal(t, ErrInvalidToken, err)
		require.Nil(t, claims)
	})
}

func TestChangePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := setupTestConfig()
	ctx := context.Background()

	mockRepo.On("FindByID", ctx, 1).Return(&User{
		ID:       1,
		Username: "testuser",
		Password: "oldhash",
	}, nil)

	mockRepo.On("FindByID", ctx, 999).Return(nil, ErrUserNotFound)

	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*auth.User")).Return(&User{}, nil).Once()
	updateErr := commonerrors.WrapError(ErrUserPasswordChangeFailed, errors.New("update error"))
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*auth.User")).Return(nil, updateErr).Once()

	authService := NewAuthService(mockRepo, cfg)

	t.Run("should_change_password_when_user_exists", func(t *testing.T) {
		err := authService.ChangePassword(ctx, 1, "newpassword")
		require.NoError(t, err)
	})

	t.Run("should_return_error_when_user_not_found", func(t *testing.T) {
		err := authService.ChangePassword(ctx, 999, "newpassword")
		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
	})

	t.Run("should_return_error_when_update_fails", func(t *testing.T) {
		err := authService.ChangePassword(ctx, 1, "newpassword")
		require.Error(t, err)
		require.Equal(t, ErrUserPasswordChangeFailed, err)
	})

	mockRepo.AssertExpectations(t)
}
