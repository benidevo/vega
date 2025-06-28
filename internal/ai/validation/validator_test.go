package validation

import (
	"errors"
	"testing"

	"github.com/benidevo/vega/internal/ai/models"
	"github.com/stretchr/testify/assert"
)

func TestAIRequestValidator_ValidateRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     models.Request
		expectError bool
	}{
		{
			name: "valid request",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Software Engineer",
				JobDescription:   "Backend Developer role",
			},
			expectError: false,
		},
		{
			name: "missing applicant name",
			request: models.Request{
				ApplicantName:    "",
				ApplicantProfile: "Software Engineer",
				JobDescription:   "Backend Developer role",
			},
			expectError: true,
		},
		{
			name: "whitespace only name",
			request: models.Request{
				ApplicantName:    "   ",
				ApplicantProfile: "Software Engineer",
				JobDescription:   "Backend Developer role",
			},
			expectError: true,
		},
		{
			name: "missing job description",
			request: models.Request{
				ApplicantName:    "John Doe",
				ApplicantProfile: "Software Engineer",
				JobDescription:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAIRequestValidator()
			err := validator.ValidateRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, models.ErrValidationFailed))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
