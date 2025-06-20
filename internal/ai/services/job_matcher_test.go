package services

import (
	"context"
	"testing"

	"github.com/benidevo/ascentio/internal/ai/llm"
	"github.com/benidevo/ascentio/internal/ai/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	// Disable logs for tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// MockProvider mocks the llm.Provider interface
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Generate(ctx context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return llm.GenerateResponse{}, args.Error(1)
	}
	return args.Get(0).(llm.GenerateResponse), args.Error(1)
}

func createTestRequest() models.Request {
	return models.Request{
		ApplicantName: "John Doe",
		ApplicantProfile: `Current Title: Senior Developer
Industry: Technology
Location: San Francisco
Career Summary: Experienced software engineer with 5+ years
Skills: Go, Python, JavaScript
Work Experience:
- Software Engineer at Previous Corp (2020 - Present)
  Built web applications`,
		JobDescription: `Position: Software Engineer
Company: Tech Corp

Description:
Build amazing software

Required Skills: Go, Python, Docker`,
		ExtraContext: "Additional context about experience",
	}
}

func TestJobMatcherService_AnalyzeMatch(t *testing.T) {
	tests := []struct {
		name           string
		request        models.Request
		setupMock      func(*MockProvider)
		expectedResult *models.MatchResult
		expectedError  error
	}{
		{
			name:    "Successful analysis",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				matchResult := models.MatchResult{
					MatchScore: 85,
					Strengths:  []string{"Strong Go skills", "Relevant experience"},
					Weaknesses: []string{"Limited Docker experience"},
					Highlights: []string{"Perfect cultural fit"},
					Feedback:   "Great candidate for this role",
				}

				response := llm.GenerateResponse{
					Data: matchResult,
				}

				provider.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeMatchResult
				})).Return(response, nil)
			},
			expectedResult: &models.MatchResult{
				MatchScore: 85,
				Strengths:  []string{"Strong Go skills", "Relevant experience"},
				Weaknesses: []string{"Limited Docker experience"},
				Highlights: []string{"Perfect cultural fit"},
				Feedback:   "Great candidate for this role",
			},
			expectedError: nil,
		},
		{
			name: "Missing applicant name",
			request: models.Request{
				ApplicantName:    "", // Missing
				ApplicantProfile: "Some profile",
				JobDescription:   "Some job description",
			},
			setupMock: func(provider *MockProvider) {
				// No mock setup needed as validation should fail
			},
			expectedResult: nil,
			expectedError:  models.ErrValidationFailed,
		},
		{
			name: "Missing applicant profile",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "", // Missing
				JobDescription:   "Some job description",
			},
			setupMock: func(provider *MockProvider) {
				// No mock setup needed as validation should fail
			},
			expectedResult: nil,
			expectedError:  models.ErrValidationFailed,
		},
		{
			name: "Missing job description",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Some profile",
				JobDescription:   "", // Missing
			},
			setupMock: func(provider *MockProvider) {
				// No mock setup needed as validation should fail
			},
			expectedResult: nil,
			expectedError:  models.ErrValidationFailed,
		},
		{
			name:    "LLM provider error",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
		{
			name:    "Invalid response type",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				// Return wrong type
				response := llm.GenerateResponse{
					Data: "invalid type",
				}

				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(response, nil)
			},
			expectedResult: nil,
			expectedError:  nil, // Will be a type assertion error
		},
		{
			name:    "Match result with invalid score gets validated",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				matchResult := models.MatchResult{
					MatchScore: 150, // Invalid score > 100
					Strengths:  []string{},
					Weaknesses: []string{},
					Highlights: []string{},
					Feedback:   "",
				}

				response := llm.GenerateResponse{
					Data: matchResult,
				}

				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(response, nil)
			},
			expectedResult: &models.MatchResult{
				MatchScore: 0, // Should be corrected to 0
				Strengths:  []string{"No specific strengths identified"},
				Weaknesses: []string{"No specific weaknesses identified"},
				Highlights: []string{"No specific highlights identified"},
				Feedback:   "Unable to provide detailed feedback at this time.",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{}
			tt.setupMock(mockProvider)

			service := NewJobMatcherService(mockProvider)

			result, err := service.AnalyzeMatch(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError != assert.AnError {
					// Check if it's a wrapped validation error
					assert.Contains(t, err.Error(), "validation failed")
				}
				assert.Nil(t, result)
			} else if tt.name == "Invalid response type" {
				// Special case: should get type assertion error
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected response type")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.MatchScore, result.MatchScore)
				assert.Equal(t, tt.expectedResult.Strengths, result.Strengths)
				assert.Equal(t, tt.expectedResult.Weaknesses, result.Weaknesses)
				assert.Equal(t, tt.expectedResult.Highlights, result.Highlights)
				assert.Equal(t, tt.expectedResult.Feedback, result.Feedback)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestJobMatcherService_validateMatchResult(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.MatchResult
		expected *models.MatchResult
	}{
		{
			name: "Valid result unchanged",
			input: &models.MatchResult{
				MatchScore: 75,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
			expected: &models.MatchResult{
				MatchScore: 75,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
		},
		{
			name: "Score too high gets corrected",
			input: &models.MatchResult{
				MatchScore: 150,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
			expected: &models.MatchResult{
				MatchScore: 0,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
		},
		{
			name: "Score too low gets corrected",
			input: &models.MatchResult{
				MatchScore: -10,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
			expected: &models.MatchResult{
				MatchScore: 0,
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Need improvement"},
				Highlights: []string{"Great experience"},
				Feedback:   "Overall good candidate",
			},
		},
		{
			name: "Empty arrays get default values",
			input: &models.MatchResult{
				MatchScore: 75,
				Strengths:  []string{},
				Weaknesses: []string{},
				Highlights: []string{},
				Feedback:   "",
			},
			expected: &models.MatchResult{
				MatchScore: 75,
				Strengths:  []string{"No specific strengths identified"},
				Weaknesses: []string{"No specific weaknesses identified"},
				Highlights: []string{"No specific highlights identified"},
				Feedback:   "Unable to provide detailed feedback at this time.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{}
			service := NewJobMatcherService(mockProvider)

			service.validateMatchResult(tt.input)

			assert.Equal(t, tt.expected.MatchScore, tt.input.MatchScore)
			assert.Equal(t, tt.expected.Strengths, tt.input.Strengths)
			assert.Equal(t, tt.expected.Weaknesses, tt.input.Weaknesses)
			assert.Equal(t, tt.expected.Highlights, tt.input.Highlights)
			assert.Equal(t, tt.expected.Feedback, tt.input.Feedback)
		})
	}
}

func TestJobMatcherService_GetMatchCategories(t *testing.T) {
	tests := []struct {
		name             string
		score            int
		expectedCategory string
		expectedDesc     string
	}{
		{
			name:             "excellent match - score 90",
			score:            90,
			expectedCategory: MatchCategoryExcellent,
			expectedDesc:     MatchDescExcellent,
		},
		{
			name:             "excellent match - score 100",
			score:            100,
			expectedCategory: MatchCategoryExcellent,
			expectedDesc:     MatchDescExcellent,
		},
		{
			name:             "strong match - score 80",
			score:            80,
			expectedCategory: MatchCategoryStrong,
			expectedDesc:     MatchDescStrong,
		},
		{
			name:             "strong match - score 89",
			score:            89,
			expectedCategory: MatchCategoryStrong,
			expectedDesc:     MatchDescStrong,
		},
		{
			name:             "good match - score 70",
			score:            70,
			expectedCategory: MatchCategoryGood,
			expectedDesc:     MatchDescGood,
		},
		{
			name:             "good match - score 79",
			score:            79,
			expectedCategory: MatchCategoryGood,
			expectedDesc:     MatchDescGood,
		},
		{
			name:             "fair match - score 60",
			score:            60,
			expectedCategory: MatchCategoryFair,
			expectedDesc:     MatchDescFair,
		},
		{
			name:             "fair match - score 69",
			score:            69,
			expectedCategory: MatchCategoryFair,
			expectedDesc:     MatchDescFair,
		},
		{
			name:             "partial match - score 50",
			score:            50,
			expectedCategory: MatchCategoryPartial,
			expectedDesc:     MatchDescPartial,
		},
		{
			name:             "partial match - score 59",
			score:            59,
			expectedCategory: MatchCategoryPartial,
			expectedDesc:     MatchDescPartial,
		},
		{
			name:             "poor match - score 49",
			score:            49,
			expectedCategory: MatchCategoryPoor,
			expectedDesc:     MatchDescPoor,
		},
		{
			name:             "poor match - score 0",
			score:            0,
			expectedCategory: MatchCategoryPoor,
			expectedDesc:     MatchDescPoor,
		},
		{
			name:             "poor match - negative score",
			score:            -10,
			expectedCategory: MatchCategoryPoor,
			expectedDesc:     MatchDescPoor,
		},
		{
			name:             "excellent match - very high score",
			score:            150,
			expectedCategory: MatchCategoryExcellent,
			expectedDesc:     MatchDescExcellent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{}
			service := NewJobMatcherService(mockProvider)

			category, desc := service.GetMatchCategories(tt.score)

			assert.Equal(t, tt.expectedCategory, category)
			assert.Equal(t, tt.expectedDesc, desc)
		})
	}
}
