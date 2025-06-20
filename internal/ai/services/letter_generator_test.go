package services

import (
	"context"
	"errors"
	"testing"

	"github.com/benidevo/ascentio/internal/ai/llm"
	"github.com/benidevo/ascentio/internal/ai/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCoverLetterGeneratorService_GenerateCoverLetter(t *testing.T) {
	tests := []struct {
		name           string
		request        models.Request
		setupMock      func(*MockProvider)
		expectedResult *models.CoverLetter
		expectedError  error
	}{
		{
			name:    "successful cover letter generation",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				coverLetter := models.CoverLetter{
					Content: "Dear Hiring Manager,\n\nI am writing to express my interest...",
					Format:  models.CoverLetterTypePlainText,
				}

				response := llm.GenerateResponse{
					Data: coverLetter,
				}

				provider.On("Generate", mock.Anything, mock.MatchedBy(func(req llm.GenerateRequest) bool {
					return req.ResponseType == llm.ResponseTypeCoverLetter
				})).Return(response, nil)
			},
			expectedResult: &models.CoverLetter{
				Content: "Dear Hiring Manager,\n\nI am writing to express my interest...",
				Format:  models.CoverLetterTypePlainText,
			},
			expectedError: nil,
		},
		{
			name: "missing applicant name",
			request: models.Request{
				ApplicantName:    "",
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
			name: "missing applicant profile",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "",
				JobDescription:   "Some job description",
			},
			setupMock: func(provider *MockProvider) {
				// No mock setup needed as validation should fail
			},
			expectedResult: nil,
			expectedError:  models.ErrValidationFailed,
		},
		{
			name: "missing job description",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Some profile",
				JobDescription:   "",
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
					Return(nil, errors.New("provider error"))
			},
			expectedResult: nil,
			expectedError:  nil, // Will be a wrapped provider error
		},
		{
			name:    "invalid response type from provider",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
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
			name:    "empty cover letter content",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				coverLetter := models.CoverLetter{
					Content: "",
					Format:  models.CoverLetterTypePlainText,
				}

				response := llm.GenerateResponse{
					Data: coverLetter,
				}

				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(response, nil)
			},
			expectedResult: nil,
			expectedError:  models.ErrValidationFailed,
		},
		{
			name:    "cover letter with empty format gets default",
			request: createTestRequest(),
			setupMock: func(provider *MockProvider) {
				coverLetter := models.CoverLetter{
					Content: "Valid cover letter content",
					Format:  "", // Empty format
				}

				response := llm.GenerateResponse{
					Data: coverLetter,
				}

				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(response, nil)
			},
			expectedResult: &models.CoverLetter{
				Content: "Valid cover letter content",
				Format:  models.CoverLetterTypePlainText, // Should default to plain text
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{}
			tt.setupMock(mockProvider)

			service := NewCoverLetterGeneratorService(mockProvider)

			result, err := service.GenerateCoverLetter(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedError == models.ErrValidationFailed {
					assert.Contains(t, err.Error(), "validation failed")
				}
			} else if tt.name == "LLM provider error" {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else if tt.name == "invalid response type from provider" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected response type")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Format, result.Format)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestCoverLetterGeneratorService_validateCoverLetter(t *testing.T) {
	tests := []struct {
		name           string
		input          *models.CoverLetter
		expectedError  bool
		expectedLetter *models.CoverLetter
	}{
		{
			name: "valid cover letter unchanged",
			input: &models.CoverLetter{
				Content: "Valid content",
				Format:  models.CoverLetterTypeHtml,
			},
			expectedError: false,
			expectedLetter: &models.CoverLetter{
				Content: "Valid content",
				Format:  models.CoverLetterTypeHtml,
			},
		},
		{
			name: "empty content returns error",
			input: &models.CoverLetter{
				Content: "",
				Format:  models.CoverLetterTypePlainText,
			},
			expectedError: true,
		},
		{
			name: "empty format gets default",
			input: &models.CoverLetter{
				Content: "Valid content",
				Format:  "",
			},
			expectedError: false,
			expectedLetter: &models.CoverLetter{
				Content: "Valid content",
				Format:  models.CoverLetterTypePlainText,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockProvider{}
			service := NewCoverLetterGeneratorService(mockProvider)

			err := service.validateCoverLetter(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, models.ErrValidationFailed)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLetter.Content, tt.input.Content)
				assert.Equal(t, tt.expectedLetter.Format, tt.input.Format)
			}
		})
	}
}
