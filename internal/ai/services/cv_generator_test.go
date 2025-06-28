package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/ai/testutil"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestCVGeneratorService_GenerateCV(t *testing.T) {
	tests := []struct {
		name          string
		request       models.Request
		setupMock     func(*testutil.MockProvider)
		expectError   bool
		errorContains string
	}{
		{
			name: "successful CV generation",
			request: models.Request{
				ApplicantName:    "Sarah Johnson",
				ApplicantProfile: "Senior Frontend Developer with 6+ years experience",
				JobDescription:   "Frontend Developer position requiring React skills",
			},
			setupMock: func(provider *testutil.MockProvider) {
				testData := testutil.NewTestData()
				result := testData.ValidGeneratedCV()
				provider.SetupCVGenerationMock(result, nil)
			},
		},
		{
			name: "empty request",
			request: models.Request{
				ApplicantName:    "",
				ApplicantProfile: "",
				JobDescription:   "",
			},
			setupMock: func(provider *testutil.MockProvider) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "provider error",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Software Engineer",
				JobDescription:   "Developer position",
			},
			setupMock: func(provider *testutil.MockProvider) {
				testData := testutil.NewTestData()
				result := testData.ValidGeneratedCV()
				provider.SetupCVGenerationMock(result, fmt.Errorf("AI service error"))
			},
			expectError:   true,
			errorContains: "AI service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &testutil.MockProvider{}
			tt.setupMock(mockProvider)

			service := NewCVGeneratorService(mockProvider)
			result, err := service.GenerateCV(context.Background(), tt.request, 1, "Test Job")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsValid)
				assert.NotEmpty(t, result.PersonalInfo.FirstName)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}
