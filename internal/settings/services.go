package settings

import (
	"context"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/logger"
	"github.com/benidevo/prospector/internal/settings/interfaces"
	"github.com/benidevo/prospector/internal/settings/models"
	"github.com/rs/zerolog"
)

// SettingsService provides business logic for user settings management
type SettingsService struct {
	settingsRepo interfaces.SettingsRepository
	cfg          *config.Settings
	log          zerolog.Logger
}

// NewSettingsService creates a new SettingsService instance
func NewSettingsService(settingsRepo interfaces.SettingsRepository, cfg *config.Settings) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
		cfg:          cfg,
		log:          logger.GetLogger("settings"),
	}
}

// GetProfileSettings retrieves a user's profile settings
func (s *SettingsService) GetProfileSettings(ctx context.Context, userID int) (*models.ProfileSettings, error) {
	return &models.ProfileSettings{
		ID:            1,
		UserID:        userID,
		FirstName:     "John",
		LastName:      "Doe",
		Title:         "Software Engineer",
		Industry:      "Technology",
		CareerSummary: "Experienced software engineer with a passion for building web applications.",
		Skills:        []string{"Go", "JavaScript", "SQL", "Docker"},
		PhoneNumber:   "555-123-4567",
		Location:      "San Francisco, CA",
	}, nil
}

// GetWorkExperiences retrieves a user's work experiences
func (s *SettingsService) GetWorkExperiences(ctx context.Context, profileID int) ([]*models.WorkExperience, error) {

	return []*models.WorkExperience{}, nil
}

// GetEducation retrieves a user's education entries
func (s *SettingsService) GetEducation(ctx context.Context, profileID int) ([]*models.Education, error) {

	return []*models.Education{}, nil
}

// GetCertifications retrieves a user's certifications
func (s *SettingsService) GetCertifications(ctx context.Context, profileID int) ([]*models.Certification, error) {

	return []*models.Certification{}, nil
}

// GetAwards retrieves a user's awards
func (s *SettingsService) GetAwards(ctx context.Context, profileID int) ([]*models.Award, error) {

	return []*models.Award{}, nil
}

// GetSecuritySettings retrieves a user's security settings
func (s *SettingsService) GetSecuritySettings(ctx context.Context, userID int) (*models.SecuritySettings, error) {

	return &models.SecuritySettings{
		ID:               1,
		UserID:           userID,
		TwoFactorEnabled: false,
		TwoFactorMethod:  "",
	}, nil
}

// GetNotificationSettings retrieves a user's notification settings
func (s *SettingsService) GetNotificationSettings(ctx context.Context, userID int) (*models.NotificationSettings, error) {

	return &models.NotificationSettings{
		ID:                 1,
		UserID:             userID,
		EmailNotifications: true,
		JobAlerts:          true,
		ApplicationUpdates: true,
		WeeklyDigest:       true,
	}, nil
}
