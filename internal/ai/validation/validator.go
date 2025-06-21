package validation

import (
	"fmt"

	"github.com/benidevo/vega/internal/ai/models"
)

// AIRequestValidator provides validation for AI service requests
type AIRequestValidator struct{}

// NewAIRequestValidator creates a new AI request validator
func NewAIRequestValidator() *AIRequestValidator {
	return &AIRequestValidator{}
}

// ValidateRequest validates the basic requirements for AI requests
func (v *AIRequestValidator) ValidateRequest(req models.Request) error {
	if req.ApplicantName == "" || req.ApplicantProfile == "" || req.JobDescription == "" {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("missing required fields: applicant name, profile, and job description are required"))
	}
	return nil
}
