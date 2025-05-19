package interfaces

import (
	"context"

	"github.com/benidevo/ascentio/internal/settings/models"
)

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	GetProfile(ctx context.Context, userID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, profile *models.Profile) error

	GetWorkExperiences(ctx context.Context, profileID int) ([]models.WorkExperience, error)
	AddWorkExperience(ctx context.Context, experience *models.WorkExperience) error
	UpdateWorkExperience(ctx context.Context, experience *models.WorkExperience) (*models.WorkExperience, error)
	DeleteWorkExperience(ctx context.Context, id int) error

	GetEducation(ctx context.Context, profileID int) ([]models.Education, error)
	AddEducation(ctx context.Context, education *models.Education) error
	UpdateEducation(ctx context.Context, education *models.Education) (*models.Education, error)
	DeleteEducation(ctx context.Context, id int) error

	GetCertifications(ctx context.Context, profileID int) ([]models.Certification, error)
	AddCertification(ctx context.Context, certification *models.Certification) error
	UpdateCertification(ctx context.Context, certification *models.Certification) (*models.Certification, error)
	DeleteCertification(ctx context.Context, id int) error
}
