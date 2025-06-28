package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/benidevo/vega/internal/ai/llm"
	"github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/ai/testutil"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestCVParserService_ParseCV(t *testing.T) {
	tests := []struct {
		name          string
		cvText        string
		setupMock     func(*testutil.MockProvider)
		expectError   bool
		errorContains string
	}{
		{
			name:   "valid CV",
			cvText: "John Doe\nSoftware Engineer\njohn@email.com\nGo, Python, React",
			setupMock: func(provider *testutil.MockProvider) {
				result := models.CVParsingResult{
					IsValid: true,
					PersonalInfo: models.PersonalInfo{
						FirstName: "John",
						LastName:  "Doe",
					},
					WorkExperience: []models.WorkExperience{},
					Education:      []models.Education{},
					Skills:         []string{"Go", "Python", "React"},
				}
				provider.SetupCVParsingMock(result, nil)
			},
		},
		{
			name:   "empty input",
			cvText: "",
			setupMock: func(provider *testutil.MockProvider) {
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name:   "invalid document",
			cvText: "This is a police report, not a CV",
			setupMock: func(provider *testutil.MockProvider) {
				provider.On("Generate", mock.Anything, mock.AnythingOfType("llm.GenerateRequest")).
					Return(llm.GenerateResponse{}, fmt.Errorf("invalid document: not a CV"))
			},
			expectError:   true,
			errorContains: "invalid document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &testutil.MockProvider{}
			tt.setupMock(mockProvider)

			service := NewCVParserService(mockProvider)
			result, err := service.ParseCV(context.Background(), tt.cvText)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsValid)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}
