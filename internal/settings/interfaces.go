package settings

import (
	"context"

	"github.com/benidevo/vega/internal/settings/models"
)

// profileService defines the methods needed for profile management
type profileService interface {
	GetProfileSettings(ctx context.Context, userID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, profile *models.Profile) error
	DeleteAllWorkExperience(ctx context.Context, profileID int) error
	DeleteAllEducation(ctx context.Context, profileID int) error
	DeleteAllCertifications(ctx context.Context, profileID int) error
}

// securityService defines the methods needed for security settings management
type securityService interface {
	GetSecuritySettings(ctx context.Context, username string) (*models.SecuritySettings, error)
}

// workExperienceService defines the methods needed for work experience management
type workExperienceService interface {
	GetWorkExperiences(ctx context.Context, profileID int) ([]models.WorkExperience, error)
	AddWorkExperience(ctx context.Context, exp *models.WorkExperience) error
	UpdateWorkExperience(ctx context.Context, exp *models.WorkExperience) (*models.WorkExperience, error)
	DeleteWorkExperience(ctx context.Context, id int, profileID int) error
}

// educationService defines the methods needed for education management
type educationService interface {
	GetEducation(ctx context.Context, profileID int) ([]models.Education, error)
	AddEducation(ctx context.Context, edu *models.Education) error
	UpdateEducation(ctx context.Context, edu *models.Education) (*models.Education, error)
	DeleteEducation(ctx context.Context, id int, profileID int) error
}

// certificationService defines the methods needed for certification management
type certificationService interface {
	GetCertifications(ctx context.Context, profileID int) ([]models.Certification, error)
	AddCertification(ctx context.Context, cert *models.Certification) error
	UpdateCertification(ctx context.Context, cert *models.Certification) (*models.Certification, error)
	DeleteCertification(ctx context.Context, id int, profileID int) error
}
