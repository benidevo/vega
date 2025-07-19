package settings

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/auth/models"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/config"
	settingsModels "github.com/benidevo/vega/internal/settings/models"
	"github.com/benidevo/vega/internal/settings/services"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	// Disable logs for tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// MockProfileRepository mocks the ProfileRepository interface
type MockProfileRepository struct {
	mock.Mock
}

// Optimized methods
func (m *MockProfileRepository) GetProfile(ctx context.Context, userID int) (*settingsModels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetProfileWithRelated(ctx context.Context, userID int) (*settingsModels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Profile), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, profile *settingsModels.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) CreateProfileIfNotExists(ctx context.Context, userID int) (*settingsModels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetEntityByID(ctx context.Context, entityID, profileID int, entityType string) (interface{}, error) {
	args := m.Called(ctx, entityID, profileID, entityType)
	return args.Get(0), args.Error(1)
}

func (m *MockProfileRepository) GetWorkExperiences(ctx context.Context, profileID int) ([]settingsModels.WorkExperience, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]settingsModels.WorkExperience), args.Error(1)
}

func (m *MockProfileRepository) AddWorkExperience(ctx context.Context, exp *settingsModels.WorkExperience) error {
	args := m.Called(ctx, exp)
	return args.Error(0)
}

func (m *MockProfileRepository) UpdateWorkExperience(ctx context.Context, exp *settingsModels.WorkExperience) (*settingsModels.WorkExperience, error) {
	args := m.Called(ctx, exp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.WorkExperience), args.Error(1)
}

func (m *MockProfileRepository) DeleteWorkExperience(ctx context.Context, id int, profileID int) error {
	args := m.Called(ctx, id, profileID)
	return args.Error(0)
}

func (m *MockProfileRepository) GetEducation(ctx context.Context, profileID int) ([]settingsModels.Education, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]settingsModels.Education), args.Error(1)
}

func (m *MockProfileRepository) AddEducation(ctx context.Context, edu *settingsModels.Education) error {
	args := m.Called(ctx, edu)
	return args.Error(0)
}

func (m *MockProfileRepository) UpdateEducation(ctx context.Context, edu *settingsModels.Education) (*settingsModels.Education, error) {
	args := m.Called(ctx, edu)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Education), args.Error(1)
}

func (m *MockProfileRepository) DeleteEducation(ctx context.Context, id int, profileID int) error {
	args := m.Called(ctx, id, profileID)
	return args.Error(0)
}

func (m *MockProfileRepository) GetCertifications(ctx context.Context, profileID int) ([]settingsModels.Certification, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]settingsModels.Certification), args.Error(1)
}

func (m *MockProfileRepository) AddCertification(ctx context.Context, cert *settingsModels.Certification) error {
	args := m.Called(ctx, cert)
	return args.Error(0)
}

func (m *MockProfileRepository) UpdateCertification(ctx context.Context, cert *settingsModels.Certification) (*settingsModels.Certification, error) {
	args := m.Called(ctx, cert)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Certification), args.Error(1)
}

func (m *MockProfileRepository) DeleteCertification(ctx context.Context, id int, profileID int) error {
	args := m.Called(ctx, id, profileID)
	return args.Error(0)
}

func (m *MockProfileRepository) DeleteAllWorkExperience(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileRepository) DeleteAllEducation(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileRepository) CreateJobSearchPreference(ctx context.Context, userID int, preference *settingsModels.JobSearchPreference) error {
	args := m.Called(ctx, userID, preference)
	return args.Error(0)
}

func (m *MockProfileRepository) GetJobSearchPreferenceByID(ctx context.Context, userID int, preferenceID string) (*settingsModels.JobSearchPreference, error) {
	args := m.Called(ctx, userID, preferenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.JobSearchPreference), args.Error(1)
}

func (m *MockProfileRepository) GetJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*settingsModels.JobSearchPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*settingsModels.JobSearchPreference), args.Error(1)
}

func (m *MockProfileRepository) GetActiveJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*settingsModels.JobSearchPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*settingsModels.JobSearchPreference), args.Error(1)
}

func (m *MockProfileRepository) UpdateJobSearchPreference(ctx context.Context, userID int, preference *settingsModels.JobSearchPreference) error {
	args := m.Called(ctx, userID, preference)
	return args.Error(0)
}

func (m *MockProfileRepository) DeleteJobSearchPreference(ctx context.Context, userID int, preferenceID string) error {
	args := m.Called(ctx, userID, preferenceID)
	return args.Error(0)
}

func (m *MockProfileRepository) ToggleJobSearchPreferenceActive(ctx context.Context, userID int, preferenceID string) error {
	args := m.Called(ctx, userID, preferenceID)
	return args.Error(0)
}

// MockUserRepository mocks the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, username, password, role string) (*models.User, error) {
	args := m.Called(ctx, username, password, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAllUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func setupTestService() (*SettingsService, *MockProfileRepository, *MockUserRepository) {
	cfg := &config.Settings{
		IsTest:   true,
		LogLevel: "disabled",
	}

	mockProfileRepo := new(MockProfileRepository)
	mockUserRepo := new(MockUserRepository)

	service := &SettingsService{
		userRepo:         mockUserRepo,
		settingsRepo:     mockProfileRepo,
		cfg:              cfg,
		log:              logger.GetPrivacyLogger("settings-test"),
		centralValidator: services.NewCentralizedValidator(),
	}
	return service, mockProfileRepo, mockUserRepo
}

func createTestProfile(userID int) *settingsModels.Profile {
	return &settingsModels.Profile{
		ID:        1,
		UserID:    userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Title:     "Software Engineer",
		Industry:  settingsModels.IndustryTechnology,
		Skills:    []string{"Go", "Python", "JavaScript"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestGetProfileSettings(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("existing profile", func(t *testing.T) {
		profile := createTestProfile(1)
		mockProfileRepo.On("GetProfileWithRelated", ctx, 1).Return(profile, nil).Once()

		result, err := service.GetProfileSettings(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, profile, result)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("profile not found - creates new", func(t *testing.T) {
		mockProfileRepo.On("GetProfileWithRelated", ctx, 2).Return(nil, nil).Once()

		result, err := service.GetProfileSettings(ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, result.UserID)
		assert.Empty(t, result.Skills)
		assert.Empty(t, result.WorkExperience)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockProfileRepo.On("GetProfileWithRelated", ctx, 3).Return(nil, dbErr).Once()

		result, err := service.GetProfileSettings(ctx, 3)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("valid profile update", func(t *testing.T) {
		profile := createTestProfile(1)
		profile.Sanitize() // Service should sanitize

		mockProfileRepo.On("UpdateProfile", ctx, profile).Return(nil).Once()

		err := service.UpdateProfile(ctx, profile)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		profile := &settingsModels.Profile{
			UserID: 0, // Invalid - missing user ID
		}

		err := service.UpdateProfile(ctx, profile)
		assert.Error(t, err)
		mockProfileRepo.AssertNotCalled(t, "UpdateProfile")
	})

	t.Run("repository error", func(t *testing.T) {
		profile := createTestProfile(1)
		repoErr := errors.New("repository error")

		mockProfileRepo.On("UpdateProfile", ctx, profile).Return(repoErr).Once()

		err := service.UpdateProfile(ctx, profile)
		assert.Error(t, err)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestGetSecuritySettings(t *testing.T) {
	ctx := context.Background()
	service, _, mockUserRepo := setupTestService()

	t.Run("user found", func(t *testing.T) {
		lastLogin := time.Now().Add(-24 * time.Hour)
		createdAt := time.Now().Add(-30 * 24 * time.Hour)

		user := &models.User{
			ID:        1,
			Username:  "johndoe",
			LastLogin: lastLogin,
			CreatedAt: createdAt,
		}

		mockUserRepo.On("FindByUsername", ctx, "johndoe").Return(user, nil).Once()

		result, err := service.GetSecuritySettings(ctx, "johndoe")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Activity)
		assert.Equal(t, lastLogin, result.Activity.LastLogin)
		assert.Equal(t, createdAt, result.Activity.CreatedAt)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.On("FindByUsername", ctx, "unknown").Return(nil, sql.ErrNoRows).Once()

		result, err := service.GetSecuritySettings(ctx, "unknown")
		assert.Error(t, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestSanitizationInService(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("profile sanitization", func(t *testing.T) {
		profile := &settingsModels.Profile{
			UserID:    1,
			FirstName: "  John  ",
			LastName:  "  Doe  ",
			Title:     "  Software Engineer  ",
			Skills:    []string{"  Go  ", "  Python  "},
		}

		mockProfileRepo.On("UpdateProfile", ctx, mock.MatchedBy(func(p *settingsModels.Profile) bool {
			return p.FirstName == "John" && p.LastName == "Doe" &&
				p.Title == "Software Engineer" && len(p.Skills) == 2 &&
				p.Skills[0] == "Go" && p.Skills[1] == "Python"
		})).Return(nil).Once()

		err := service.UpdateProfile(ctx, profile)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

}

func TestDeleteAllWorkExperience(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("successful deletion", func(t *testing.T) {
		profileID := 1
		mockProfileRepo.On("DeleteAllWorkExperience", ctx, profileID).Return(nil).Once()

		err := service.DeleteAllWorkExperience(ctx, profileID)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		profileID := 1
		repoErr := errors.New("database error")
		mockProfileRepo.On("DeleteAllWorkExperience", ctx, profileID).Return(repoErr).Once()

		err := service.DeleteAllWorkExperience(ctx, profileID)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestDeleteAllEducation(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("successful deletion", func(t *testing.T) {
		profileID := 1
		mockProfileRepo.On("DeleteAllEducation", ctx, profileID).Return(nil).Once()

		err := service.DeleteAllEducation(ctx, profileID)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		profileID := 1
		repoErr := errors.New("database error")
		mockProfileRepo.On("DeleteAllEducation", ctx, profileID).Return(repoErr).Once()

		err := service.DeleteAllEducation(ctx, profileID)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestSettingsService_CreatePreference(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("successful creation", func(t *testing.T) {
		userID := 1
		input := services.CreatePreferenceInput{
			JobTitle: "Backend Engineer",
			Location: "San Francisco",
			Skills:   []string{"Go", "Kubernetes", "PostgreSQL"},
			MaxAge:   settingsModels.MaxAgeOneWeek,
			IsActive: true,
		}

		// Mock that user has no existing preferences
		mockProfileRepo.On("GetJobSearchPreferencesByUserID", ctx, userID).
			Return([]*settingsModels.JobSearchPreference{}, nil).Once()

		mockProfileRepo.On("CreateJobSearchPreference", ctx, userID, mock.AnythingOfType("*models.JobSearchPreference")).
			Run(func(args mock.Arguments) {
				// Simulate what the real repository does
				pref := args.Get(2).(*settingsModels.JobSearchPreference)
				pref.ID = "generated-uuid"
				pref.UserID = userID
				pref.CreatedAt = time.Now()
				pref.UpdatedAt = time.Now()
			}).
			Return(nil).Once()

		pref, err := service.CreatePreference(ctx, userID, input)
		require.NoError(t, err)
		assert.NotNil(t, pref)
		assert.Equal(t, input.JobTitle, pref.JobTitle)
		assert.Equal(t, input.Location, pref.Location)
		assert.Equal(t, input.Skills, pref.Skills)
		assert.Equal(t, input.MaxAge, pref.MaxAge)
		assert.Equal(t, input.IsActive, pref.IsActive)
		assert.NotEmpty(t, pref.ID) // Should have generated UUID

		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("max preferences reached", func(t *testing.T) {
		userID := 1
		input := services.CreatePreferenceInput{
			JobTitle: "Backend Engineer",
			Location: "San Francisco",
			MaxAge:   settingsModels.MaxAgeOneWeek,
			IsActive: true,
		}

		// Mock that user already has 4 preferences (the limit)
		existingPrefs := make([]*settingsModels.JobSearchPreference, services.MaxPreferencesPerUser)
		for i := 0; i < services.MaxPreferencesPerUser; i++ {
			existingPrefs[i] = &settingsModels.JobSearchPreference{ID: "test-id"}
		}

		mockProfileRepo.On("GetJobSearchPreferencesByUserID", ctx, userID).
			Return(existingPrefs, nil).Once()

		pref, err := service.CreatePreference(ctx, userID, input)
		assert.Nil(t, pref)
		assert.Equal(t, services.ErrMaxPreferencesReached, err)

		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("invalid max age", func(t *testing.T) {
		userID := 1
		input := services.CreatePreferenceInput{
			JobTitle: "Backend Engineer",
			Location: "San Francisco",
			MaxAge:   100, // Too low
			IsActive: true,
		}

		pref, err := service.CreatePreference(ctx, userID, input)
		assert.Nil(t, pref)
		assert.Equal(t, services.ErrInvalidPreferenceData, err)

		mockProfileRepo.AssertNotCalled(t, "GetJobSearchPreferencesByUserID")
	})
}

func TestSettingsService_GetUserPreferences(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("successful retrieval", func(t *testing.T) {
		userID := 1
		expectedPrefs := []*settingsModels.JobSearchPreference{
			{
				ID:       "pref-1",
				UserID:   userID,
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   settingsModels.MaxAgeOneDay,
				IsActive: true,
			},
			{
				ID:       "pref-2",
				UserID:   userID,
				JobTitle: "Frontend Developer",
				Location: "New York",
				MaxAge:   settingsModels.MaxAgeOneWeek,
				IsActive: false,
			},
		}

		mockProfileRepo.On("GetJobSearchPreferencesByUserID", ctx, userID).
			Return(expectedPrefs, nil).Once()

		prefs, err := service.GetUserPreferences(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedPrefs, prefs)

		mockProfileRepo.AssertExpectations(t)
	})
}

func TestSettingsService_DeletePreference(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	t.Run("successful deletion", func(t *testing.T) {
		userID := 1
		prefID := "test-pref-id"

		mockProfileRepo.On("DeleteJobSearchPreference", ctx, userID, prefID).
			Return(nil).Once()

		err := service.DeletePreference(ctx, userID, prefID)
		assert.NoError(t, err)

		mockProfileRepo.AssertExpectations(t)
	})
}
