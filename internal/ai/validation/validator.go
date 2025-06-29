package validation

import (
	"fmt"
	"strings"

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
	if err := v.validateBasicFields(req); err != nil {
		return err
	}

	if err := v.validateFieldLengths(req); err != nil {
		return err
	}

	if err := v.validateContent(req); err != nil {
		return err
	}

	return nil
}

// validateBasicFields checks for required fields
func (v *AIRequestValidator) validateBasicFields(req models.Request) error {
	if v.isEmpty(req.ApplicantName) {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("applicant name is required"))
	}

	if v.isEmpty(req.ApplicantProfile) && v.isEmpty(req.CVText) {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("either applicant profile or CV text must be provided"))
	}

	if v.isEmpty(req.JobDescription) {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("job description is required"))
	}

	return nil
}

// validateFieldLengths checks field length constraints
func (v *AIRequestValidator) validateFieldLengths(req models.Request) error {
	const (
		maxNameLength         = 200
		maxProfileLength      = 50000 // ~10 page CV
		maxJobDescLength      = 20000 // Large job descriptions
		maxExtraContextLength = 5000
		maxCVTextLength       = 50000 // Same as profile
	)

	if len(req.ApplicantName) > maxNameLength {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("applicant name too long (max %d characters)", maxNameLength))
	}

	if len(req.ApplicantProfile) > maxProfileLength {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("applicant profile too long (max %d characters)", maxProfileLength))
	}

	if len(req.JobDescription) > maxJobDescLength {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("job description too long (max %d characters)", maxJobDescLength))
	}

	if len(req.ExtraContext) > maxExtraContextLength {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("extra context too long (max %d characters)", maxExtraContextLength))
	}

	if len(req.CVText) > maxCVTextLength {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("CV text too long (max %d characters)", maxCVTextLength))
	}

	return nil
}

// validateContent performs basic content validation
func (v *AIRequestValidator) validateContent(req models.Request) error {
	if v.containsSuspiciousContent(req.ApplicantProfile) ||
		v.containsSuspiciousContent(req.CVText) ||
		v.containsSuspiciousContent(req.JobDescription) ||
		v.containsSuspiciousContent(req.ExtraContext) {
		return models.WrapError(models.ErrValidationFailed,
			fmt.Errorf("content contains potentially unsafe or invalid data"))
	}

	return nil
}

// containsSuspiciousContent checks for obviously problematic content
func (v *AIRequestValidator) containsSuspiciousContent(content string) bool {
	if content == "" {
		return false
	}

	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"data:text/html",
		"<?php",
		"<iframe",
		"eval(",
		"window.location",
		"document.cookie",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}

	return false
}

// isEmpty checks if a string is empty or contains only whitespace
func (v *AIRequestValidator) isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
