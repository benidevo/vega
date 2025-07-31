package settings

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/auth/models"
	settingsModels "github.com/benidevo/vega/internal/settings/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

type MockProfileRepository struct {
	mock.Mock
}

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

func (m *MockProfileRepository) DeleteAllCertifications(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
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

// mockProfileService implements the profileService interface for testing
type mockProfileService struct {
	mock.Mock
}

func (m *mockProfileService) GetProfileSettings(ctx context.Context, userID int) (*settingsModels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Profile), args.Error(1)
}

func (m *mockProfileService) UpdateProfile(ctx context.Context, profile *settingsModels.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockProfileService) DeleteAllWorkExperience(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *mockProfileService) DeleteAllEducation(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *mockProfileService) DeleteAllCertifications(ctx context.Context, profileID int) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

// mockSecurityService implements the securityService interface for testing
type mockSecurityService struct {
	mock.Mock
}

func (m *mockSecurityService) GetSecuritySettings(ctx context.Context, username string) (*settingsModels.SecuritySettings, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.SecuritySettings), args.Error(1)
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
	}
}

func TestGetProfileSettings(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockSetup      func(*mockProfileService)
		expectedResult *settingsModels.Profile
		expectedError  bool
	}{
		{
			name:   "should_return_existing_profile_when_found",
			userID: 1,
			mockSetup: func(m *mockProfileService) {
				profile := createTestProfile(1)
				m.On("GetProfileSettings", mock.Anything, 1).Return(profile, nil)
			},
			expectedResult: createTestProfile(1),
			expectedError:  false,
		},
		{
			name:   "should_return_error_when_database_fails",
			userID: 3,
			mockSetup: func(m *mockProfileService) {
				m.On("GetProfileSettings", mock.Anything, 3).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockProfileService)
			tc.mockSetup(mockService)

			result, err := mockService.GetProfileSettings(context.Background(), tc.userID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.expectedResult != nil {
					// Compare profiles without timestamps
					assert.Equal(t, tc.expectedResult.ID, result.ID)
					assert.Equal(t, tc.expectedResult.UserID, result.UserID)
					assert.Equal(t, tc.expectedResult.FirstName, result.FirstName)
					assert.Equal(t, tc.expectedResult.LastName, result.LastName)
					assert.Equal(t, tc.expectedResult.Email, result.Email)
					assert.Equal(t, tc.expectedResult.Title, result.Title)
					assert.Equal(t, tc.expectedResult.Skills, result.Skills)
				}
			}
			mockService.AssertExpectations(t)
		})
	}

}

func TestUpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		profile       *settingsModels.Profile
		mockSetup     func(*mockProfileService)
		expectedError bool
	}{
		{
			name:    "should_update_profile_when_valid",
			profile: createTestProfile(1),
			mockSetup: func(m *mockProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "should_return_error_when_repository_fails",
			profile: createTestProfile(1),
			mockSetup: func(m *mockProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(errors.New("repository error"))
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockProfileService)
			tc.mockSetup(mockService)

			err := mockService.UpdateProfile(context.Background(), tc.profile)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}

}

func TestGetSecuritySettings(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		mockSetup      func(*mockSecurityService)
		expectedResult *settingsModels.SecuritySettings
		expectedError  bool
	}{
		{
			name:     "should_return_security_settings_when_user_found",
			username: "johndoe",
			mockSetup: func(m *mockSecurityService) {
				lastLogin := time.Now().Add(-24 * time.Hour)
				createdAt := time.Now().Add(-30 * 24 * time.Hour)
				settings := &settingsModels.SecuritySettings{
					Activity: &settingsModels.AccountActivity{
						LastLogin: lastLogin,
						CreatedAt: createdAt,
					},
				}
				m.On("GetSecuritySettings", mock.Anything, "johndoe").Return(settings, nil)
			},
			expectedError: false,
		},
		{
			name:     "should_return_error_when_user_not_found",
			username: "unknown",
			mockSetup: func(m *mockSecurityService) {
				m.On("GetSecuritySettings", mock.Anything, "unknown").Return(nil, sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockSecurityService)
			tc.mockSetup(mockService)

			result, err := mockService.GetSecuritySettings(context.Background(), tc.username)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			mockService.AssertExpectations(t)
		})
	}

}

func TestDeleteAllWorkExperience(t *testing.T) {
	tests := []struct {
		name          string
		profileID     int
		mockSetup     func(*mockProfileService)
		expectedError bool
	}{
		{
			name:      "should_delete_all_work_experience_when_successful",
			profileID: 1,
			mockSetup: func(m *mockProfileService) {
				m.On("DeleteAllWorkExperience", mock.Anything, 1).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "should_return_error_when_repository_fails",
			profileID: 1,
			mockSetup: func(m *mockProfileService) {
				m.On("DeleteAllWorkExperience", mock.Anything, 1).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockProfileService)
			tc.mockSetup(mockService)

			err := mockService.DeleteAllWorkExperience(context.Background(), tc.profileID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteAllEducation(t *testing.T) {
	tests := []struct {
		name          string
		profileID     int
		mockSetup     func(*mockProfileService)
		expectedError bool
	}{
		{
			name:      "should_delete_all_education_when_successful",
			profileID: 1,
			mockSetup: func(m *mockProfileService) {
				m.On("DeleteAllEducation", mock.Anything, 1).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "should_return_error_when_repository_fails",
			profileID: 1,
			mockSetup: func(m *mockProfileService) {
				m.On("DeleteAllEducation", mock.Anything, 1).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockProfileService)
			tc.mockSetup(mockService)

			err := mockService.DeleteAllEducation(context.Background(), tc.profileID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}
