package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(username, password, role string) (*User, error) {
	args := m.Called(username, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *User) (*User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAllUsers() ([]*User, error) {
	args := m.Called()
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
	t.Run("should successfully register a new user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("CreateUser", "testuser", mock.AnythingOfType("string"), "admin").Return(&User{
			ID:       1,
			Username: "testuser",
			Role:     ADMIN,
		}, nil)

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.Register("testuser", "password123", "admin")

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, "testuser", user.Username)
		require.Equal(t, ADMIN, user.Role)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("CreateUser", "testuser", mock.AnythingOfType("string"), "admin").Return(nil, errors.New("repository error"))

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.Register("testuser", "password123", "admin")

		require.Error(t, err)
		require.Equal(t, ErrUserCreationFailed, err)
		require.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	t.Run("should successfully login a user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		hashedPassword, err := hashPassword("password123")
		require.NoError(t, err)

		mockRepo.On("FindByUsername", "testuser").Return(&User{
			ID:       1,
			Username: "testuser",
			Password: hashedPassword,
			Role:     ADMIN,
		}, nil)

		mockRepo.On("UpdateUser", mock.AnythingOfType("*auth.User")).Return(&User{}, nil)

		authService := NewAuthService(mockRepo, cfg)

		token, err := authService.Login("testuser", "password123")

		require.NoError(t, err)
		require.NotEmpty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByUsername", "testuser").Return(nil, ErrUserNotFound)

		authService := NewAuthService(mockRepo, cfg)

		token, err := authService.Login("testuser", "password123")

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Empty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when password is incorrect", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		hashedPassword, err := hashPassword("correctpassword")
		require.NoError(t, err)

		mockRepo.On("FindByUsername", "testuser").Return(&User{
			ID:       1,
			Username: "testuser",
			Password: hashedPassword,
			Role:     ADMIN,
		}, nil)

		authService := NewAuthService(mockRepo, cfg)

		token, err := authService.Login("testuser", "wrongpassword")

		require.Error(t, err)
		require.Equal(t, ErrInvalidCredentials, err)
		require.Empty(t, token)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("should return user when found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByID", 1).Return(&User{
			ID:       1,
			Username: "testuser",
			Role:     ADMIN,
		}, nil)

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.GetUserByID(1)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, 1, user.ID)
		require.Equal(t, "testuser", user.Username)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByID", 999).Return(nil, ErrUserNotFound)

		authService := NewAuthService(mockRepo, cfg)

		user, err := authService.GetUserByID(999)

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		require.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}

func TestVerifyToken(t *testing.T) {
	t.Run("should verify valid token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		authService := NewAuthService(mockRepo, cfg)

		user := &User{
			ID:       1,
			Username: "testuser",
			Role:     ADMIN,
		}

		token, err := authService.GenerateToken(user)
		require.NoError(t, err)

		claims, err := authService.VerifyToken(token)

		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, 1, claims.UserID)
		require.Equal(t, "testuser", claims.Username)
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		authService := NewAuthService(mockRepo, cfg)

		claims, err := authService.VerifyToken("invalid.token.here")

		require.Error(t, err)
		require.Equal(t, ErrInvalidToken, err)
		require.Nil(t, claims)
	})

	t.Run("should reject expired token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)

		expiredCfg := &config.Settings{
			TokenSecret:     "test-secret-key",
			TokenExpiration: -1 * time.Hour, // Expired 1 hour ago
		}

		authService := NewAuthService(mockRepo, expiredCfg)

		user := &User{
			ID:       1,
			Username: "testuser",
			Role:     ADMIN,
		}

		token, err := authService.GenerateToken(user)
		require.NoError(t, err)

		authService = NewAuthService(mockRepo, setupTestConfig())

		claims, err := authService.VerifyToken(token)

		require.Error(t, err)
		require.Nil(t, claims)
	})
}

func TestChangePassword(t *testing.T) {
	t.Run("should change password successfully", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByID", 1).Return(&User{
			ID:       1,
			Username: "testuser",
			Password: "oldhash",
		}, nil)

		mockRepo.On("UpdateUser", mock.AnythingOfType("*auth.User")).Return(&User{}, nil)

		authService := NewAuthService(mockRepo, cfg)

		err := authService.ChangePassword(1, "newpassword")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByID", 999).Return(nil, ErrUserNotFound)

		authService := NewAuthService(mockRepo, cfg)

		err := authService.ChangePassword(999, "newpassword")

		require.Error(t, err)
		require.Equal(t, ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		cfg := setupTestConfig()

		mockRepo.On("FindByID", 1).Return(&User{
			ID:       1,
			Username: "testuser",
			Password: "oldhash",
		}, nil)

		mockRepo.On("UpdateUser", mock.AnythingOfType("*auth.User")).Return(nil, errors.New("update error"))

		authService := NewAuthService(mockRepo, cfg)

		err := authService.ChangePassword(1, "newpassword")

		require.Error(t, err)
		require.Equal(t, ErrUserPasswordChangeFailed, err)
		mockRepo.AssertExpectations(t)
	})
}
