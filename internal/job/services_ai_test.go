package job

import (
	"context"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/ai"
	aimodels "github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Disable logs for tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// MockJobMatcherService mocks the JobMatcherService
type MockJobMatcherService struct {
	mock.Mock
}

func (m *MockJobMatcherService) AnalyzeMatch(ctx context.Context, req aimodels.Request) (*aimodels.MatchResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aimodels.MatchResult), args.Error(1)
}

func (m *MockJobMatcherService) GetMatchCategories(score int) (string, string) {
	args := m.Called(score)
	return args.String(0), args.String(1)
}

// MockCoverLetterGeneratorService mocks the CoverLetterGeneratorService
type MockCoverLetterGeneratorService struct {
	mock.Mock
}

func (m *MockCoverLetterGeneratorService) GenerateCoverLetter(ctx context.Context, req aimodels.Request) (*aimodels.CoverLetter, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aimodels.CoverLetter), args.Error(1)
}

// MockSettingsService mocks the SettingsService
type MockSettingsService struct {
	mock.Mock
}

func (m *MockSettingsService) GetProfileSettings(ctx context.Context, userID int) (*settingsmodels.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*settingsmodels.Profile), args.Error(1)
}

func createTestJobForAI() *models.Job {
	return &models.Job{
		ID:             1,
		Title:          "Software Engineer",
		Description:    "Build amazing software",
		RequiredSkills: []string{"Go", "Python", "Docker"},
		Company: models.Company{
			ID:   1,
			Name: "Tech Corp",
		},
		Status:    models.INTERESTED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestProfile() *settingsmodels.Profile {
	return &settingsmodels.Profile{
		FirstName:     "John",
		LastName:      "Doe",
		Title:         "Senior Developer",
		Industry:      settingsmodels.IndustryTechnology,
		Location:      "San Francisco",
		CareerSummary: "Experienced software engineer with 5+ years",
		Skills:        []string{"Go", "Python", "JavaScript"},
		WorkExperience: []settingsmodels.WorkExperience{
			{
				Title:       "Software Engineer",
				Company:     "Previous Corp",
				StartDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Description: "Built web applications",
			},
		},
		Education: []settingsmodels.Education{
			{
				Degree:       "BS",
				FieldOfStudy: "Computer Science",
				Institution:  "Tech University",
				StartDate:    time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		Context: "Additional context about my experience",
	}
}

func TestJobService_AnalyzeJobMatch_NoAIService(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}

	service := NewJobService(mockJobRepo, nil, nil, cfg)

	result, err := service.AnalyzeJobMatch(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Equal(t, models.ErrAIServiceUnavailable, err)
	assert.Nil(t, result)
}

func TestJobService_AnalyzeJobMatch_NoSettingsService(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}

	// Create service with AI service but no settings service
	service := NewJobService(mockJobRepo, &ai.AIService{}, nil, cfg)

	result, err := service.AnalyzeJobMatch(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Equal(t, models.ErrProfileServiceRequired, err)
	assert.Nil(t, result)
}

func TestJobService_GenerateCoverLetter_NoAIService(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}

	service := NewJobService(mockJobRepo, nil, nil, cfg)

	result, err := service.GenerateCoverLetter(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Equal(t, models.ErrAIServiceUnavailable, err)
	assert.Nil(t, result)
}

func TestJobService_ValidateProfileForAI(t *testing.T) {
	tests := []struct {
		name          string
		profile       *settingsmodels.Profile
		expectedError error
	}{
		{
			name:          "Valid profile",
			profile:       createTestProfile(),
			expectedError: nil,
		},
		{
			name: "Missing name",
			profile: &settingsmodels.Profile{
				Skills:        []string{"Go", "Python"},
				CareerSummary: "Experienced developer",
			},
			expectedError: models.ErrProfileIncomplete,
		},
		{
			name: "Missing career info",
			profile: &settingsmodels.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Skills:    []string{"Go", "Python"},
			},
			expectedError: models.ErrProfileSummaryRequired,
		},
		{
			name: "Valid with work experience but no career summary",
			profile: &settingsmodels.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Skills:    []string{"Go", "Python"},
				WorkExperience: []settingsmodels.WorkExperience{
					{
						Title:   "Software Engineer",
						Company: "Tech Corp",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Invalid work experience without career summary",
			profile: &settingsmodels.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Skills:    []string{"Go", "Python"},
				WorkExperience: []settingsmodels.WorkExperience{
					{
						Title:   "", // Empty title
						Company: "",
					},
				},
			},
			expectedError: models.ErrProfileSummaryRequired,
		},
		{
			name: "Valid with education only",
			profile: &settingsmodels.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Skills:    []string{"Go", "Python"},
				Education: []settingsmodels.Education{
					{
						Degree:       "BS",
						FieldOfStudy: "Computer Science",
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockJobRepo := &MockJobRepository{}
			cfg := &config.Settings{}
			service := NewJobService(mockJobRepo, nil, nil, cfg)

			err := service.ValidateProfileForAI(tt.profile)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJobService_buildAIRequest(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	job := createTestJobForAI()
	profile := createTestProfile()

	request := service.buildAIRequest(job, profile)

	assert.Equal(t, "John Doe", request.ApplicantName)
	assert.Contains(t, request.ApplicantProfile, "Senior Developer")
	assert.Contains(t, request.ApplicantProfile, "Technology")
	assert.Contains(t, request.ApplicantProfile, "Go, Python, JavaScript")
	assert.Contains(t, request.JobDescription, "Software Engineer")
	assert.Contains(t, request.JobDescription, "Tech Corp")
	assert.Contains(t, request.JobDescription, "Go, Python, Docker")
	assert.Equal(t, "Additional context about my experience", request.ExtraContext)
}

func TestJobService_buildAIRequest_EmptyName(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	job := createTestJobForAI()
	profile := createTestProfile()
	profile.FirstName = ""
	profile.LastName = ""

	request := service.buildAIRequest(job, profile)

	assert.Equal(t, "Applicant", request.ApplicantName)
}

func TestJobService_buildProfileSummary(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	profile := createTestProfile()

	summary := service.buildProfileSummary(profile)

	assert.Contains(t, summary, "Current Title: Senior Developer")
	assert.Contains(t, summary, "Industry: Technology")
	assert.Contains(t, summary, "Location: San Francisco")
	assert.Contains(t, summary, "Career Summary:")
	assert.Contains(t, summary, "Skills: Go, Python, JavaScript")
	assert.Contains(t, summary, "Work Experience:")
	assert.Contains(t, summary, "Education:")
	assert.Contains(t, summary, "Additional Context:")
}

func TestJobService_buildProfileSummary_LimitedExperience(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	profile := createTestProfile()

	// Add many work experiences to test the limit
	for i := 0; i < 10; i++ {
		profile.WorkExperience = append(profile.WorkExperience, settingsmodels.WorkExperience{
			Title:       "Developer",
			Company:     "Company",
			StartDate:   time.Date(2020-i, 1, 1, 0, 0, 0, 0, time.UTC),
			Description: "Some description that is quite long and should be truncated if it exceeds the character limit for descriptions in the profile summary generation process",
		})
	}

	summary := service.buildProfileSummary(profile)

	// Should limit to 5 experiences
	experiences := len(profile.WorkExperience)
	assert.True(t, experiences > 5)
	// Count occurrences of " at " which indicates work experience entries
	// Should be 5 (limited) + context about limiting
	assert.Contains(t, summary, "Work Experience:")
}

func TestJobService_convertToJobMatchAnalysis(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	aiResult := &aimodels.MatchResult{
		MatchScore: 85,
		Strengths:  []string{"Strong Go skills"},
		Weaknesses: []string{"Limited Docker experience"},
		Highlights: []string{"Perfect fit"},
		Feedback:   "Great candidate",
	}

	result := service.convertToJobMatchAnalysis(aiResult, 1, 2)

	assert.Equal(t, 2, result.JobID)
	assert.Equal(t, 1, result.UserID)
	assert.Equal(t, 85, result.MatchScore)
	assert.Equal(t, []string{"Strong Go skills"}, result.Strengths)
	assert.Equal(t, []string{"Limited Docker experience"}, result.Weaknesses)
	assert.Equal(t, []string{"Perfect fit"}, result.Highlights)
	assert.Equal(t, "Great candidate", result.Feedback)
	assert.NotZero(t, result.AnalyzedAt)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)
}

func TestJobService_convertToCoverLetter(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}
	service := NewJobService(mockJobRepo, nil, nil, cfg)

	aiResult := &aimodels.CoverLetter{
		Content: "Dear Hiring Manager...",
		Format:  aimodels.CoverLetterTypeMarkdown,
	}

	result := service.convertToCoverLetter(aiResult, 1, 2)

	assert.Equal(t, 2, result.JobID)
	assert.Equal(t, 1, result.UserID)
	assert.Equal(t, "Dear Hiring Manager...", result.Content)
	assert.Equal(t, string(aimodels.CoverLetterTypeMarkdown), result.Format)
	assert.NotZero(t, result.GeneratedAt)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)
}

func TestJobService_GenerateCV_NoAIService(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}

	service := NewJobService(mockJobRepo, nil, nil, cfg)

	result, err := service.GenerateCV(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Equal(t, models.ErrAIServiceUnavailable, err)
	assert.Nil(t, result)
}

func TestJobService_GenerateCV_NoSettingsService(t *testing.T) {
	mockJobRepo := &MockJobRepository{}
	cfg := &config.Settings{}

	service := NewJobService(mockJobRepo, &ai.AIService{}, nil, cfg)

	result, err := service.GenerateCV(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Equal(t, models.ErrProfileServiceRequired, err)
	assert.Nil(t, result)
}
