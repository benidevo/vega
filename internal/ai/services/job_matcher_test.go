package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/benidevo/vega/internal/ai/llm"
	"github.com/benidevo/vega/internal/ai/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockJobMatcher struct {
	mock.Mock
}

func (m *MockJobMatcher) Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error) {
	args := m.Called(ctx, request)
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
		setupMock      func(*MockJobMatcher)
		expectedResult *models.MatchResult
		expectError    bool
		errorContains  string
	}{
		{
			name:    "should_match_job_when_request_valid",
			request: createTestRequest(),
			setupMock: func(m *MockJobMatcher) {
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

				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
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
		},
		{
			name: "should_return_error_when_applicant_name_missing",
			request: models.Request{
				ApplicantName:    "", // Missing
				ApplicantProfile: "Some profile",
				JobDescription:   "Some job description",
			},
			setupMock: func(m *MockJobMatcher) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "should_return_error_when_applicant_profile_empty",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "", // Empty
				JobDescription:   "Some job description",
			},
			setupMock: func(m *MockJobMatcher) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "should_return_error_when_job_description_empty",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Some profile",
				JobDescription:   "", // Empty
			},
			setupMock: func(m *MockJobMatcher) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name:    "should_return_error_when_provider_fails",
			request: createTestRequest(),
			setupMock: func(m *MockJobMatcher) {
				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeMatchResult
				})).Return(llm.GenerateResponse{}, fmt.Errorf("AI service error"))
			},
			expectError:   true,
			errorContains: "AI service error",
		},
		{
			name:    "should_return_error_when_response_invalid_type",
			request: createTestRequest(),
			setupMock: func(m *MockJobMatcher) {
				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeMatchResult
				})).Return(llm.GenerateResponse{
					Data: "invalid type",
				}, nil)
			},
			expectError:   true,
			errorContains: "unexpected response type",
		},
		{
			name: "should_include_previous_matches_when_provided",
			request: models.Request{
				ApplicantName: "Jane Smith",
				ApplicantProfile: `Current Title: Data Scientist
Skills: Python, ML, SQL`,
				JobDescription: "Data Scientist role requiring ML experience",
				PreviousMatches: []models.PreviousMatch{
					{
						JobTitle:    "ML Engineer",
						Company:     "AI Corp",
						MatchScore:  75,
						KeyInsights: "Good ML skills but limited production experience",
						DaysAgo:     5,
					},
				},
			},
			setupMock: func(m *MockJobMatcher) {
				matchResult := models.MatchResult{
					MatchScore: 80,
					Strengths:  []string{"Strong ML background"},
					Weaknesses: []string{"Limited production experience"},
					Highlights: []string{"Excellent Python skills"},
					Feedback:   "Good fit with room for growth",
				}

				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeMatchResult
				})).Return(llm.GenerateResponse{
					Data: matchResult,
				}, nil)
			},
			expectedResult: &models.MatchResult{
				MatchScore: 80,
				Strengths:  []string{"Strong ML background"},
				Weaknesses: []string{"Limited production experience"},
				Highlights: []string{"Excellent Python skills"},
				Feedback:   "Good fit with room for growth",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockJobMatcher{}
			if tt.setupMock != nil {
				tt.setupMock(mockProvider)
			}

			service := NewJobMatcherService(mockProvider)
			result, err := service.AnalyzeMatch(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestJobMatcherService_GetMatchCategories(t *testing.T) {
	tests := []struct {
		name                string
		score               int
		expectedCategory    string
		expectedDescription string
	}{
		{
			name:                "should_return_excellent_when_score_90_or_above",
			score:               95,
			expectedCategory:    "Excellent Match",
			expectedDescription: "You are an outstanding fit for the role with minimal gaps.",
		},
		{
			name:                "should_return_strong_when_score_80_to_89",
			score:               85,
			expectedCategory:    "Strong Match",
			expectedDescription: "You have strong qualifications with only minor areas for development.",
		},
		{
			name:                "should_return_good_when_score_70_to_79",
			score:               75,
			expectedCategory:    "Good Match",
			expectedDescription: "You meet most requirements with some skill gaps that can be addressed.",
		},
		{
			name:                "should_return_fair_when_score_60_to_69",
			score:               65,
			expectedCategory:    "Fair Match",
			expectedDescription: "You have potential but may need significant development in key areas.",
		},
		{
			name:                "should_return_partial_when_score_50_to_59",
			score:               55,
			expectedCategory:    "Partial Match",
			expectedDescription: "You have some relevant qualifications but significant gaps exist.",
		},
		{
			name:                "should_return_poor_when_score_below_50",
			score:               30,
			expectedCategory:    "Poor Match",
			expectedDescription: "You do not meet the core requirements for this position.",
		},
		{
			name:                "should_handle_boundary_score_90",
			score:               90,
			expectedCategory:    "Excellent Match",
			expectedDescription: "You are an outstanding fit for the role with minimal gaps.",
		},
		{
			name:                "should_handle_boundary_score_80",
			score:               80,
			expectedCategory:    "Strong Match",
			expectedDescription: "You have strong qualifications with only minor areas for development.",
		},
		{
			name:                "should_handle_boundary_score_70",
			score:               70,
			expectedCategory:    "Good Match",
			expectedDescription: "You meet most requirements with some skill gaps that can be addressed.",
		},
		{
			name:                "should_handle_boundary_score_60",
			score:               60,
			expectedCategory:    "Fair Match",
			expectedDescription: "You have potential but may need significant development in key areas.",
		},
		{
			name:                "should_handle_boundary_score_50",
			score:               50,
			expectedCategory:    "Partial Match",
			expectedDescription: "You have some relevant qualifications but significant gaps exist.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockJobMatcher{}
			service := NewJobMatcherService(mockProvider)

			category, description := service.GetMatchCategories(tt.score)

			assert.Equal(t, tt.expectedCategory, category)
			assert.Equal(t, tt.expectedDescription, description)
		})
	}
}
