package models

import (
	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
)

var (
	// Model validation errors
	ErrInvalidSettingsType = commonerrors.New("invalid settings type")
	ErrFieldRequired       = commonerrors.New("field is required")
	ErrInvalidFieldParam   = commonerrors.New("invalid field parameter")

	// Special validation error that can't be handled by struct tags
	ErrEndDateWithCurrent = commonerrors.New("end date must be empty when position is current")

	// Repository errors
	ErrSettingsNotFound       = commonerrors.New("settings not found")
	ErrFailedToGetSettings    = commonerrors.New("failed to get settings")
	ErrFailedToUpdateSettings = commonerrors.New("failed to update settings")

	ErrProfileNotFound = commonerrors.New("profile not found")

	// Specific entity not found errors
	ErrWorkExperienceNotFound = commonerrors.New("work experience not found")
	ErrEducationNotFound      = commonerrors.New("education not found")
	ErrCertificationNotFound  = commonerrors.New("certification not found")

	// Service layer errors
	ErrFailedToCreateWorkExperience = commonerrors.New("failed to create work experience")
	ErrFailedToUpdateWorkExperience = commonerrors.New("failed to update work experience")
	ErrFailedToDeleteWorkExperience = commonerrors.New("failed to delete work experience")
	ErrFailedToGetWorkExperience    = commonerrors.New("failed to get work experience")

	ErrFailedToCreateEducation = commonerrors.New("failed to create education")
	ErrFailedToUpdateEducation = commonerrors.New("failed to update education")
	ErrFailedToDeleteEducation = commonerrors.New("failed to delete education")
	ErrFailedToGetEducation    = commonerrors.New("failed to get education")

	ErrFailedToCreateCertification = commonerrors.New("failed to create certification")
	ErrFailedToUpdateCertification = commonerrors.New("failed to update certification")
	ErrFailedToDeleteCertification = commonerrors.New("failed to delete certification")
	ErrFailedToGetCertification    = commonerrors.New("failed to get certification")
)

// WrapError is a convenience function that calls the common errors package
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}

// GetSentinelError is a convenience function that calls the common errors package
func GetSentinelError(err error) error {
	return commonerrors.GetSentinelError(err)
}
