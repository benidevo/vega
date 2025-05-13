package interfaces

import (
	"context"

	"github.com/benidevo/prospector/internal/settings/models"
)

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	GetProfileSettings(ctx context.Context, userID int) (*models.ProfileSettings, error)
	UpdateProfileSettings(ctx context.Context, profile *models.ProfileSettings) error

	GetWorkExperiences(ctx context.Context, profileID int) ([]*models.WorkExperience, error)
	AddWorkExperience(ctx context.Context, experience *models.WorkExperience) error
	UpdateWorkExperience(ctx context.Context, experience *models.WorkExperience) error
	DeleteWorkExperience(ctx context.Context, id int) error

	GetEducation(ctx context.Context, profileID int) ([]*models.Education, error)
	AddEducation(ctx context.Context, education *models.Education) error
	UpdateEducation(ctx context.Context, education *models.Education) error
	DeleteEducation(ctx context.Context, id int) error

	GetCertifications(ctx context.Context, profileID int) ([]*models.Certification, error)
	AddCertification(ctx context.Context, certification *models.Certification) error
	UpdateCertification(ctx context.Context, certification *models.Certification) error
	DeleteCertification(ctx context.Context, id int) error

	GetAwards(ctx context.Context, profileID int) ([]*models.Award, error)
	AddAward(ctx context.Context, award *models.Award) error
	UpdateAward(ctx context.Context, award *models.Award) error
	DeleteAward(ctx context.Context, id int) error

	GetSecuritySettings(ctx context.Context, userID int) (*models.SecuritySettings, error)
	UpdateSecuritySettings(ctx context.Context, security *models.SecuritySettings) error

	GetNotificationSettings(ctx context.Context, userID int) (*models.NotificationSettings, error)
	UpdateNotificationSettings(ctx context.Context, notifications *models.NotificationSettings) error
}
