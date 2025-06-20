package services

import (
	"context"
	"errors"
	"testing"

	"github.com/benidevo/vega/internal/auth/models"
	commonerrors "github.com/benidevo/vega/internal/common/errors"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// setupGoogleAuthTest prepares common test dependencies
func setupGoogleAuthTest() (*MockUserRepository, *config.Settings, context.Context) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Settings{
		TokenSecret:             "test-secret-key",
		GoogleAuthUserInfoURL:   "https://www.googleapis.com/oauth2/v1/userinfo",
		GoogleAuthUserInfoScope: "https://www.googleapis.com/auth/userinfo.email",
	}
	ctx := context.Background()
	return mockRepo, cfg, ctx
}

func TestGetOrCreateUser(t *testing.T) {
	mockRepo, cfg, ctx := setupGoogleAuthTest()

	log := logger.GetPrivacyLogger("test")

	service := &GoogleAuthService{
		cfg:  cfg,
		repo: mockRepo,
		log:  log,
	}

	userInfo := &GoogleAuthUserInfo{
		ID:            "12345",
		Email:         "test@example.com",
		VerifiedEmail: true,
	}

	t.Run("should_return_existing_user", func(t *testing.T) {
		existingUser := &models.User{
			ID:       1,
			Username: "test@example.com",
			Role:     models.STANDARD,
		}
		mockRepo.On("FindByUsername", ctx, "test@example.com").Return(existingUser, nil).Once()

		user, err := service.getOrCreateUser(ctx, userInfo)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, existingUser.ID, user.ID)
		require.Equal(t, existingUser.Username, user.Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should_create_new_user_when_not_found", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, "test@example.com").Return(nil, models.ErrUserNotFound).Once()

		newUser := &models.User{
			ID:       2,
			Username: "test@example.com",
			Role:     models.STANDARD,
		}
		mockRepo.On("CreateUser", ctx, "test@example.com", "", "Standard").Return(newUser, nil).Once()

		user, err := service.getOrCreateUser(ctx, userInfo)

		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, newUser.ID, user.ID)
		require.Equal(t, newUser.Username, user.Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should_return_error_when_user_lookup_fails", func(t *testing.T) {
		repoErr := commonerrors.WrapError(models.ErrUserRetrievalFailed, errors.New("database error"))
		mockRepo.On("FindByUsername", ctx, "test@example.com").Return(nil, repoErr).Once()

		user, err := service.getOrCreateUser(ctx, userInfo)

		require.Error(t, err)
		require.Equal(t, models.ErrGoogleUserCreationFailed, err)
		require.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should_return_error_when_user_creation_fails", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, "test@example.com").Return(nil, models.ErrUserNotFound).Once()

		repoErr := commonerrors.WrapError(models.ErrUserCreationFailed, errors.New("database error"))
		mockRepo.On("CreateUser", ctx, "test@example.com", "", "Standard").Return(nil, repoErr).Once()

		user, err := service.getOrCreateUser(ctx, userInfo)

		require.Error(t, err)
		require.Equal(t, models.ErrGoogleUserCreationFailed, err)
		require.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAuthURL(t *testing.T) {
	_, cfg, _ := setupGoogleAuthTest()
	log := logger.GetPrivacyLogger("test")

	t.Run("should_return_auth_url_with_state", func(t *testing.T) {
		service := &GoogleAuthService{
			cfg: cfg,
			log: log,
			oauthCfg: &oauth2.Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost/auth/google/callback",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
			},
		}

		url := service.GetAuthURL()

		require.Contains(t, url, "https://accounts.google.com/o/oauth2/auth")
		require.Contains(t, url, "client_id=test-client-id")
		require.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%2Fauth%2Fgoogle%2Fcallback")
		require.Contains(t, url, "access_type=offline")
	})
}

func TestGenerateRandomState(t *testing.T) {
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		state := generateRandomState(length)

		require.Equal(t, length, len(state))

		for _, char := range state {
			require.True(t,
				(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9'),
				"Character '%c' is not alphanumeric", char)
		}

		anotherState := generateRandomState(length)
		require.NotEqual(t, state, anotherState, "Two generated states should differ")
	}
}
