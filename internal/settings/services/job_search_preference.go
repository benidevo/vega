package services

import (
	"errors"

	"github.com/benidevo/vega/internal/settings/models"
)

// Constants for service layer
const (
	MaxPreferencesPerUser = 4 // Limit to prevent abuse
)

// JobSearchPreferenceService errors
var (
	ErrPreferenceNotFound    = errors.New("job search preference not found")
	ErrMaxPreferencesReached = errors.New("maximum number of preferences reached")
	ErrInvalidPreferenceData = errors.New("invalid preference data")
)

// CreatePreferenceInput represents the input for creating a new preference
type CreatePreferenceInput struct {
	JobTitle string   `json:"job_title" form:"job_title" binding:"required,min=1,max=100"`
	Location string   `json:"location" form:"location" binding:"required,min=1,max=100"`
	Skills   []string `json:"skills,omitempty" form:"skills" binding:"max=10,dive,min=1,max=50"`
	MaxAge   int      `json:"max_age" form:"max_age" binding:"required"`
	IsActive bool     `json:"is_active" form:"is_active"`
}

// UpdatePreferenceInput represents the input for updating a preference
type UpdatePreferenceInput struct {
	JobTitle string   `json:"job_title" form:"job_title" binding:"required,min=1,max=100"`
	Location string   `json:"location" form:"location" binding:"required,min=1,max=100"`
	Skills   []string `json:"skills,omitempty" form:"skills" binding:"max=10,dive,min=1,max=50"`
	MaxAge   int      `json:"max_age" form:"max_age" binding:"required"`
	IsActive bool     `json:"is_active" form:"is_active"`
}

// ValidateMaxAge validates that the max age is within allowed bounds
func ValidateMaxAge(maxAge int) error {
	if maxAge < models.MinMaxAge || maxAge > models.MaxMaxAge {
		return ErrInvalidPreferenceData
	}
	return nil
}
