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

type MockLetterGenerator struct {
	mock.Mock
}

func (m *MockLetterGenerator) Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return llm.GenerateResponse{}, args.Error(1)
	}
	return args.Get(0).(llm.GenerateResponse), args.Error(1)
}

func TestCoverLetterGeneratorService_GenerateCoverLetter(t *testing.T) {
	tests := []struct {
		name          string
		request       models.Request
		setupMock     func(*MockLetterGenerator)
		expectError   bool
		errorContains string
	}{
		{
			name:    "should_generate_cover_letter_when_request_valid",
			request: createTestRequest(),
			setupMock: func(m *MockLetterGenerator) {
				coverLetter := models.CoverLetter{
					Content: "Dear Hiring Manager,\n\nI am writing to express my interest...",
					Format:  models.CoverLetterTypePlainText,
				}

				response := llm.GenerateResponse{
					Data: coverLetter,
				}

				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeCoverLetter
				})).Return(response, nil)
			},
		},
		{
			name: "should_return_error_when_applicant_name_missing",
			request: models.Request{
				ApplicantName:    "",
				ApplicantProfile: "Some profile",
				JobDescription:   "Some job description",
			},
			setupMock: func(m *MockLetterGenerator) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "should_return_error_when_applicant_profile_missing",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "",
				JobDescription:   "Some job description",
			},
			setupMock: func(m *MockLetterGenerator) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "should_return_error_when_job_description_missing",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Some profile",
				JobDescription:   "",
			},
			setupMock: func(m *MockLetterGenerator) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name:    "should_return_error_when_provider_fails",
			request: createTestRequest(),
			setupMock: func(m *MockLetterGenerator) {
				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeCoverLetter
				})).Return(llm.GenerateResponse{}, fmt.Errorf("AI service error"))
			},
			expectError:   true,
			errorContains: "AI service error",
		},
		{
			name:    "should_return_error_when_response_invalid_type",
			request: createTestRequest(),
			setupMock: func(m *MockLetterGenerator) {
				response := llm.GenerateResponse{
					Data: "invalid type",
				}

				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeCoverLetter
				})).Return(response, nil)
			},
			expectError:   true,
			errorContains: "unexpected response type",
		},
		{
			name: "should_generate_html_format_when_configured",
			request: models.Request{
				ApplicantName:    "Jane Smith",
				ApplicantProfile: "Backend Developer with Go expertise",
				JobDescription:   "Go Developer position",
				ExtraContext:     "Focus on microservices experience",
			},
			setupMock: func(m *MockLetterGenerator) {
				coverLetter := models.CoverLetter{
					Content: "<p>Dear Hiring Manager,</p><p>I am excited to apply...</p>",
					Format:  models.CoverLetterTypeHtml,
				}

				m.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeCoverLetter
				})).Return(llm.GenerateResponse{
					Data: coverLetter,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockLetterGenerator{}
			if tt.setupMock != nil {
				tt.setupMock(mockProvider)
			}

			service := NewCoverLetterGeneratorService(mockProvider)
			result, err := service.GenerateCoverLetter(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Content)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}
