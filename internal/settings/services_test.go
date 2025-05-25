package settings

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/benidevo/ascentio/internal/auth/models"
	"github.com/benidevo/ascentio/internal/config"
	settingsModels "github.com/benidevo/ascentio/internal/settings/models"
	"github.com/go-playground/validator/v10"
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

func (m *MockProfileRepository) GetProfile(ctx context.Context, userID int) (*settingsModels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsModels.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetWorkExperiences(ctx context.Context, profileID int) ([]settingsModels.WorkExperience, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]settingsModels.WorkExperience), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, profile *settingsModels.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
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

func (m *MockProfileRepository) DeleteWorkExperience(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
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

func (m *MockProfileRepository) DeleteEducation(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
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

func (m *MockProfileRepository) DeleteCertification(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
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
		userRepo:     mockUserRepo,
		settingsRepo: mockProfileRepo,
		cfg:          cfg,
		log:          zerolog.Nop(),
		validator:    validator.New(),
	}
	return service, mockProfileRepo, mockUserRepo
}

func createTestProfile(userID int) *settingsModels.Profile {
	return &settingsModels.Profile{
		ID:        1,
		UserID:    userID,
		FirstName: "John",
		LastName:  "Doe",
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
		mockProfileRepo.On("GetProfile", ctx, 1).Return(profile, nil).Once()

		result, err := service.GetProfileSettings(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, profile, result)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("profile not found - creates new", func(t *testing.T) {
		mockProfileRepo.On("GetProfile", ctx, 2).Return(nil, nil).Once()

		result, err := service.GetProfileSettings(ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, result.UserID)
		assert.Empty(t, result.Skills)
		assert.Empty(t, result.WorkExperience)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockProfileRepo.On("GetProfile", ctx, 3).Return(nil, dbErr).Once()

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

func TestWorkExperienceOperations(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	now := time.Now()
	exp := &settingsModels.WorkExperience{
		ProfileID:   1,
		Company:     "Acme Corp",
		Title:       "Senior Developer",
		Location:    "Remote",
		StartDate:   now.Add(-365 * 24 * time.Hour),
		Description: "Building great software",
		Current:     true,
	}

	t.Run("create work experience", func(t *testing.T) {
		mockProfileRepo.On("AddWorkExperience", ctx, exp).Return(nil).Once()

		err := service.CreateWorkExperience(ctx, exp)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("create with validation error", func(t *testing.T) {
		invalidExp := &settingsModels.WorkExperience{
			ProfileID: 0, // Invalid
			Company:   "Acme Corp",
			Title:     "Senior Developer",
			StartDate: now,
		}

		err := service.CreateWorkExperience(ctx, invalidExp)
		assert.Error(t, err)
		mockProfileRepo.AssertNotCalled(t, "AddWorkExperience")
	})

	t.Run("update work experience", func(t *testing.T) {
		mockProfileRepo.On("UpdateWorkExperience", ctx, exp).Return(exp, nil).Once()

		err := service.UpdateWorkExperience(ctx, exp)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("delete work experience", func(t *testing.T) {
		// Mock GetWorkExperiences to return the experience we want to delete
		expWithID := *exp
		expWithID.ID = 10
		mockProfileRepo.On("GetWorkExperiences", ctx, 1).Return([]settingsModels.WorkExperience{expWithID}, nil).Once()
		mockProfileRepo.On("DeleteWorkExperience", ctx, 10).Return(nil).Once()

		err := service.DeleteWorkExperience(ctx, 10, 1)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestEducationOperations(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	now := time.Now()
	endDate := now.Add(-180 * 24 * time.Hour)
	edu := &settingsModels.Education{
		ProfileID:    1,
		Institution:  "MIT",
		Degree:       "BS Computer Science",
		FieldOfStudy: "Computer Science",
		StartDate:    now.Add(-4 * 365 * 24 * time.Hour),
		EndDate:      &endDate,
		Description:  "Graduated with honors",
	}

	t.Run("create education", func(t *testing.T) {
		mockProfileRepo.On("AddEducation", ctx, edu).Return(nil).Once()

		err := service.CreateEducation(ctx, edu)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("create with validation error", func(t *testing.T) {
		invalidEdu := &settingsModels.Education{
			ProfileID:   1,
			Institution: "", // Invalid - required
			Degree:      "BS Computer Science",
			StartDate:   now,
		}

		err := service.CreateEducation(ctx, invalidEdu)
		assert.Error(t, err)
		mockProfileRepo.AssertNotCalled(t, "AddEducation")
	})

	t.Run("update education", func(t *testing.T) {
		mockProfileRepo.On("UpdateEducation", ctx, edu).Return(edu, nil).Once()

		err := service.UpdateEducation(ctx, edu)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("delete education", func(t *testing.T) {
		mockProfileRepo.On("DeleteEducation", ctx, 20).Return(nil).Once()

		err := service.DeleteEducation(ctx, 20, 1)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestCertificationOperations(t *testing.T) {
	ctx := context.Background()
	service, mockProfileRepo, _ := setupTestService()

	now := time.Now()
	expiryDate := now.Add(365 * 24 * time.Hour)
	cert := &settingsModels.Certification{
		ProfileID:     1,
		Name:          "AWS Solutions Architect",
		IssuingOrg:    "Amazon Web Services",
		IssueDate:     now.Add(-180 * 24 * time.Hour),
		ExpiryDate:    &expiryDate,
		CredentialID:  "AWS-123456",
		CredentialURL: "https://aws.amazon.com/verify/AWS-123456",
	}

	t.Run("create certification", func(t *testing.T) {
		mockProfileRepo.On("AddCertification", ctx, cert).Return(nil).Once()

		err := service.CreateCertification(ctx, cert)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("create with validation error", func(t *testing.T) {
		invalidCert := &settingsModels.Certification{
			ProfileID:  1,
			Name:       "AWS Solutions Architect",
			IssuingOrg: "", // Invalid - required
			IssueDate:  now,
		}

		err := service.CreateCertification(ctx, invalidCert)
		assert.Error(t, err)
		mockProfileRepo.AssertNotCalled(t, "AddCertification")
	})

	t.Run("update certification", func(t *testing.T) {
		mockProfileRepo.On("UpdateCertification", ctx, cert).Return(cert, nil).Once()

		err := service.UpdateCertification(ctx, cert)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("delete certification", func(t *testing.T) {
		// Mock GetCertifications to return the certification we want to delete
		certWithID := *cert
		certWithID.ID = 30
		mockProfileRepo.On("GetCertifications", ctx, 1).Return([]settingsModels.Certification{certWithID}, nil).Once()
		mockProfileRepo.On("DeleteCertification", ctx, 30).Return(nil).Once()

		err := service.DeleteCertification(ctx, 30, 1)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
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

	t.Run("work experience sanitization", func(t *testing.T) {
		exp := &settingsModels.WorkExperience{
			ProfileID: 1,
			Company:   "  Acme Corp  ",
			Title:     "  Senior Developer  ",
			Location:  "  Remote  ",
			StartDate: time.Now().Add(-365 * 24 * time.Hour),
		}

		mockProfileRepo.On("AddWorkExperience", ctx, mock.MatchedBy(func(e *settingsModels.WorkExperience) bool {
			return e.Company == "Acme Corp" && e.Title == "Senior Developer" && e.Location == "Remote"
		})).Return(nil).Once()

		err := service.CreateWorkExperience(ctx, exp)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})
}
