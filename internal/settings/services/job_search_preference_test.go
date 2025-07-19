package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/benidevo/vega/internal/settings/models"
	"github.com/benidevo/vega/internal/settings/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repository for testing
type MockPreferenceRepository struct {
	mock.Mock
}

func (m *MockPreferenceRepository) CreateJobSearchPreference(ctx context.Context, userID int, preference *models.JobSearchPreference) error {
	args := m.Called(ctx, userID, preference)
	return args.Error(0)
}

func (m *MockPreferenceRepository) GetJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*models.JobSearchPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSearchPreference), args.Error(1)
}

func (m *MockPreferenceRepository) GetJobSearchPreferenceByID(ctx context.Context, userID int, preferenceID string) (*models.JobSearchPreference, error) {
	args := m.Called(ctx, userID, preferenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobSearchPreference), args.Error(1)
}

func (m *MockPreferenceRepository) UpdateJobSearchPreference(ctx context.Context, userID int, preference *models.JobSearchPreference) error {
	args := m.Called(ctx, userID, preference)
	return args.Error(0)
}

func (m *MockPreferenceRepository) DeleteJobSearchPreference(ctx context.Context, userID int, preferenceID string) error {
	args := m.Called(ctx, userID, preferenceID)
	return args.Error(0)
}

func (m *MockPreferenceRepository) ToggleJobSearchPreferenceActive(ctx context.Context, userID int, preferenceID string) error {
	args := m.Called(ctx, userID, preferenceID)
	return args.Error(0)
}

func (m *MockPreferenceRepository) GetActiveJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*models.JobSearchPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobSearchPreference), args.Error(1)
}

// Implement remaining required methods as stubs
func (m *MockPreferenceRepository) GetProfile(ctx context.Context, userID int) (*models.Profile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) GetProfileWithRelated(ctx context.Context, userID int) (*models.Profile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) CreateProfileIfNotExists(ctx context.Context, userID int) (*models.Profile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) GetEntityByID(ctx context.Context, entityID, profileID int, entityType string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) GetWorkExperiences(ctx context.Context, profileID int) ([]models.WorkExperience, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) AddWorkExperience(ctx context.Context, experience *models.WorkExperience) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) UpdateWorkExperience(ctx context.Context, experience *models.WorkExperience) (*models.WorkExperience, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) DeleteWorkExperience(ctx context.Context, id int, profileID int) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) DeleteAllWorkExperience(ctx context.Context, profileID int) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) GetEducation(ctx context.Context, profileID int) ([]models.Education, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) AddEducation(ctx context.Context, education *models.Education) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) UpdateEducation(ctx context.Context, education *models.Education) (*models.Education, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) DeleteEducation(ctx context.Context, id int, profileID int) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) DeleteAllEducation(ctx context.Context, profileID int) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) GetCertifications(ctx context.Context, profileID int) ([]models.Certification, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) AddCertification(ctx context.Context, certification *models.Certification) error {
	return errors.New("not implemented")
}

func (m *MockPreferenceRepository) UpdateCertification(ctx context.Context, certification *models.Certification) (*models.Certification, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPreferenceRepository) DeleteCertification(ctx context.Context, id int, profileID int) error {
	return errors.New("not implemented")
}

func TestCreatePreferenceInput_Validation(t *testing.T) {
	tests := []struct {
		name       string
		input      services.CreatePreferenceInput
		wantMaxErr bool // Expecting max age error specifically
	}{
		{
			name: "valid input",
			input: services.CreatePreferenceInput{
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   86400, // 1 day
				IsActive: true,
			},
			wantMaxErr: false,
		},
		{
			name: "max age too low",
			input: services.CreatePreferenceInput{
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   1800, // 30 minutes - below minimum
				IsActive: true,
			},
			wantMaxErr: true,
		},
		{
			name: "max age too high",
			input: services.CreatePreferenceInput{
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   9999999, // way above maximum
				IsActive: true,
			},
			wantMaxErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateMaxAge(tt.input.MaxAge)
			if tt.wantMaxErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaxPreferencesLimit(t *testing.T) {
	// Test that the max preferences limit is enforced at 4
	assert.Equal(t, 4, services.MaxPreferencesPerUser)
}

func TestValidateMaxAge(t *testing.T) {
	tests := []struct {
		name    string
		maxAge  int
		wantErr bool
	}{
		{"valid - 1 hour", models.MaxAgeOneHour, false},
		{"valid - 6 hours", models.MaxAgeSixHours, false},
		{"valid - 12 hours", models.MaxAgeTwelveHours, false},
		{"valid - 1 day", models.MaxAgeOneDay, false},
		{"valid - 3 days", models.MaxAgeThreeDays, false},
		{"valid - 1 week", models.MaxAgeOneWeek, false},
		{"valid - 2 weeks", models.MaxAgeTwoWeeks, false},
		{"valid - 30 days", models.MaxAgeThirtyDays, false},
		{"invalid - too low", 1800, true},     // 30 minutes
		{"invalid - too high", 9999999, true}, // way too high
		{"invalid - zero", 0, true},
		{"invalid - negative", -3600, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.ValidateMaxAge(tt.maxAge)
			if tt.wantErr {
				assert.Equal(t, services.ErrInvalidPreferenceData, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJobSearchPreference_GetMaxAgeDisplay(t *testing.T) {
	tests := []struct {
		name     string
		maxAge   int
		expected string
	}{
		{"1 hour", models.MaxAgeOneHour, "1 hour"},
		{"6 hours", models.MaxAgeSixHours, "6 hours"},
		{"12 hours", models.MaxAgeTwelveHours, "12 hours"},
		{"1 day", models.MaxAgeOneDay, "1 day"},
		{"3 days", models.MaxAgeThreeDays, "3 days"},
		{"1 week", models.MaxAgeOneWeek, "1 week"},
		{"2 weeks", models.MaxAgeTwoWeeks, "2 weeks"},
		{"30 days", models.MaxAgeThirtyDays, "30 days"},
		{"unknown value", 12345, "3h25m45s"}, // time.Duration format
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pref := &models.JobSearchPreference{
				MaxAge: tt.maxAge,
			}
			assert.Equal(t, tt.expected, pref.GetMaxAgeDisplay())
		})
	}
}

func TestJobSearchPreference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pref    *models.JobSearchPreference
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid preference",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "valid preference with skills",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				Skills:   []string{"Go", "Python", "Docker"},
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "missing job title",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "", // empty - will fail validation
				Location: "Remote",
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "JobTitle",
		},
		{
			name: "missing location",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "", // empty - will fail validation
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "Location",
		},
		{
			name: "job title too long",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: string(make([]byte, 101)), // 101 characters
				Location: "Remote",
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "JobTitle",
		},
		{
			name: "location too long",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: string(make([]byte, 101)), // 101 characters
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "Location",
		},
		{
			name: "invalid max age - too low",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   500, // too low
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "MaxAge",
		},
		{
			name: "too many skills",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				Skills:   []string{"Skill1", "Skill2", "Skill3", "Skill4", "Skill5", "Skill6", "Skill7", "Skill8", "Skill9", "Skill10", "Skill11"},
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "Skills",
		},
		{
			name: "skill too long",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				Skills:   []string{"Go", "This is a really long skill name that exceeds the maximum allowed length of 50 characters"},
				MaxAge:   models.MaxAgeOneDay,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "Skills",
		},
		{
			name: "invalid max age - too high",
			pref: &models.JobSearchPreference{
				ID:       "test-id",
				UserID:   1,
				JobTitle: "Software Engineer",
				Location: "Remote",
				MaxAge:   9999999, // too high
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "MaxAge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pref.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJobSearchPreference_Sanitize(t *testing.T) {
	pref := &models.JobSearchPreference{
		JobTitle: "  Software Engineer  ",
		Location: "  New York  ",
		Skills:   []string{"  Go  ", "  Python  ", "", "  ", "Docker"},
	}

	pref.Sanitize()

	assert.Equal(t, "Software Engineer", pref.JobTitle)
	assert.Equal(t, "New York", pref.Location)
	assert.Equal(t, []string{"Go", "Python", "Docker"}, pref.Skills)
}
