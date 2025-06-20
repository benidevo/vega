package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/benidevo/ascentio/internal/settings/models"
)

// ValidationRule represents a validation rule with its error message
type ValidationRule struct {
	Check   func() bool
	Message string
}

// CentralizedValidator provides unified validation for all entities
type CentralizedValidator struct{}

// NewCentralizedValidator creates a new centralized validator
func NewCentralizedValidator() *CentralizedValidator {
	return &CentralizedValidator{}
}

// ValidateProfile performs comprehensive validation on a profile
func (v *CentralizedValidator) ValidateProfile(profile *models.Profile) error {
	rules := []ValidationRule{
		{
			Check:   func() bool { return profile.UserID > 0 },
			Message: "User ID must be positive",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(profile.FirstName)) > 0 },
			Message: "First name is required",
		},
		{
			Check:   func() bool { return len(profile.FirstName) <= 100 },
			Message: "First name must not exceed 100 characters",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(profile.LastName)) > 0 },
			Message: "Last name is required",
		},
		{
			Check:   func() bool { return len(profile.LastName) <= 100 },
			Message: "Last name must not exceed 100 characters",
		},
		{
			Check:   func() bool { return len(profile.Title) <= 200 },
			Message: "Title must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(profile.CareerSummary) <= 2000 },
			Message: "Career summary must not exceed 2000 characters",
		},
		{
			Check:   func() bool { return len(profile.Skills) <= 50 },
			Message: "Skills must not exceed 50 items",
		},
		{
			Check:   func() bool { return len(profile.Location) <= 200 },
			Message: "Location must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(profile.PhoneNumber) <= 20 },
			Message: "Phone number must not exceed 20 characters",
		},
		{
			Check:   func() bool { return len(profile.LinkedInProfile) <= 500 },
			Message: "LinkedIn profile URL must not exceed 500 characters",
		},
		{
			Check:   func() bool { return len(profile.GitHubProfile) <= 500 },
			Message: "GitHub profile URL must not exceed 500 characters",
		},
		{
			Check:   func() bool { return len(profile.Website) <= 500 },
			Message: "Website URL must not exceed 500 characters",
		},
		{
			Check:   func() bool { return len(profile.Context) <= 6000 },
			Message: "Context must not exceed 6000 characters",
		},
	}

	// Validate skills individually
	for i, skill := range profile.Skills {
		skill := strings.TrimSpace(skill)
		if len(skill) == 0 {
			return fmt.Errorf("skill %d cannot be empty", i+1)
		}
		if len(skill) > 100 {
			return fmt.Errorf("skill %d must not exceed 100 characters", i+1)
		}
	}

	if profile.PhoneNumber != "" {
		if !v.isValidPhoneNumber(profile.PhoneNumber) {
			return fmt.Errorf("phone number contains invalid characters or must be between 10-20 characters")
		}
	}

	if profile.LinkedInProfile != "" && !v.isValidLinkedInURL(profile.LinkedInProfile) {
		return fmt.Errorf("LinkedIn profile must be a valid LinkedIn URL")
	}

	if profile.GitHubProfile != "" && !v.isValidGitHubURL(profile.GitHubProfile) {
		return fmt.Errorf("GitHub profile must be a valid GitHub URL")
	}

	if profile.Website != "" && !v.isValidURL(profile.Website) {
		return fmt.Errorf("website must be a valid URL")
	}

	for _, rule := range rules {
		if !rule.Check() {
			return fmt.Errorf("%s", rule.Message)
		}
	}

	return nil
}

// ValidateWorkExperience performs comprehensive validation on work experience
func (v *CentralizedValidator) ValidateWorkExperience(we *models.WorkExperience) error {
	rules := []ValidationRule{
		{
			Check:   func() bool { return we.ProfileID > 0 },
			Message: "Profile ID must be positive",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(we.Company)) > 0 },
			Message: "Company name is required",
		},
		{
			Check:   func() bool { return len(we.Company) <= 200 },
			Message: "Company name must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(we.Title)) > 0 },
			Message: "Job title is required",
		},
		{
			Check:   func() bool { return len(we.Title) <= 200 },
			Message: "Job title must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(we.Location) <= 200 },
			Message: "Location must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(we.Description) <= 2000 },
			Message: "Description must not exceed 2000 characters",
		},
		{
			Check:   func() bool { return !we.StartDate.After(time.Now()) },
			Message: "Start date cannot be in the future",
		},
		{
			Check:   func() bool { return !we.Current || we.EndDate == nil },
			Message: "End date must be empty when position is current",
		},
	}

	if we.EndDate != nil {
		if we.EndDate.After(time.Now()) {
			return fmt.Errorf("end date cannot be in the future")
		}
		if we.EndDate.Before(we.StartDate) {
			return fmt.Errorf("end date must be after start date")
		}
	}

	for _, rule := range rules {
		if !rule.Check() {
			return fmt.Errorf("%s", rule.Message)
		}
	}

	return nil
}

// ValidateEducation performs comprehensive validation on education
func (v *CentralizedValidator) ValidateEducation(ed *models.Education) error {
	rules := []ValidationRule{
		{
			Check:   func() bool { return ed.ProfileID > 0 },
			Message: "Profile ID must be positive",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(ed.Institution)) > 0 },
			Message: "Institution is required",
		},
		{
			Check:   func() bool { return len(ed.Institution) <= 200 },
			Message: "Institution name must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(ed.Degree)) > 0 },
			Message: "Degree is required",
		},
		{
			Check:   func() bool { return len(ed.Degree) <= 200 },
			Message: "Degree must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(ed.FieldOfStudy) <= 200 },
			Message: "Field of study must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(ed.Description) <= 2000 },
			Message: "Description must not exceed 2000 characters",
		},
		{
			Check:   func() bool { return !ed.StartDate.After(time.Now()) },
			Message: "Start date cannot be in the future",
		},
	}

	if ed.EndDate != nil {
		if ed.EndDate.After(time.Now()) {
			return fmt.Errorf("end date cannot be in the future")
		}
		if ed.EndDate.Before(ed.StartDate) {
			return fmt.Errorf("end date must be after start date")
		}
	}

	for _, rule := range rules {
		if !rule.Check() {
			return fmt.Errorf("%s", rule.Message)
		}
	}

	return nil
}

// ValidateCertification performs comprehensive validation on certification
func (v *CentralizedValidator) ValidateCertification(cert *models.Certification) error {
	rules := []ValidationRule{
		{
			Check:   func() bool { return cert.ProfileID > 0 },
			Message: "Profile ID must be positive",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(cert.Name)) > 0 },
			Message: "Certification name is required",
		},
		{
			Check:   func() bool { return len(cert.Name) <= 200 },
			Message: "Certification name must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(strings.TrimSpace(cert.IssuingOrg)) > 0 },
			Message: "Issuing organization is required",
		},
		{
			Check:   func() bool { return len(cert.IssuingOrg) <= 200 },
			Message: "Issuing organization must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(cert.CredentialID) <= 200 },
			Message: "Credential ID must not exceed 200 characters",
		},
		{
			Check:   func() bool { return len(cert.CredentialURL) <= 500 },
			Message: "Credential URL must not exceed 500 characters",
		},
		{
			Check:   func() bool { return !cert.IssueDate.After(time.Now()) },
			Message: "Issue date cannot be in the future",
		},
	}

	if cert.CredentialURL != "" && !v.isValidURL(cert.CredentialURL) {
		return fmt.Errorf("credential URL must be a valid URL")
	}

	if cert.ExpiryDate != nil {
		if cert.ExpiryDate.Before(cert.IssueDate) {
			return fmt.Errorf("expiry date must be after issue date")
		}
	}

	for _, rule := range rules {
		if !rule.Check() {
			return fmt.Errorf("%s", rule.Message)
		}
	}

	return nil
}

// ValidateWordCount validates word count for text fields
func (v *CentralizedValidator) ValidateWordCount(text string, maxWords int, fieldName string) error {
	words := strings.Fields(text)
	if len(words) > maxWords {
		return fmt.Errorf("%s must not exceed %d words", fieldName, maxWords)
	}
	return nil
}

// Helper validation methods
func (v *CentralizedValidator) isValidPhoneNumber(phone string) bool {
	if phone == "" {
		return true // Optional field
	}
	for _, r := range phone {
		if !strings.ContainsRune("0123456789 -+()", r) {
			return false
		}
	}
	return len(phone) >= 10 && len(phone) <= 20
}

func (v *CentralizedValidator) isValidLinkedInURL(url string) bool {
	if url == "" {
		return true // Optional field
	}
	return strings.HasPrefix(url, "https://www.linkedin.com/") ||
		strings.HasPrefix(url, "https://linkedin.com/") ||
		strings.HasPrefix(url, "http://www.linkedin.com/") ||
		strings.HasPrefix(url, "http://linkedin.com/")
}

func (v *CentralizedValidator) isValidGitHubURL(url string) bool {
	if url == "" {
		return true // Optional field
	}
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "https://www.github.com/") ||
		strings.HasPrefix(url, "http://github.com/") ||
		strings.HasPrefix(url, "http://www.github.com/")
}

func (v *CentralizedValidator) isValidURL(url string) bool {
	if url == "" {
		return true // Optional field
	}
	// Basic URL validation - starts with http:// or https://
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
