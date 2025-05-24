package settings

import (
	"context"
	"errors"

	authrepo "github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/settings/interfaces"
	"github.com/benidevo/ascentio/internal/settings/models"
	"github.com/rs/zerolog"
)

// SettingsService provides business logic for user settings management
type SettingsService struct {
	userRepo     authrepo.UserRepository
	settingsRepo interfaces.SettingsRepository
	cfg          *config.Settings
	log          zerolog.Logger
}

// NewSettingsService creates a new SettingsService instance
func NewSettingsService(settingsRepo interfaces.SettingsRepository, cfg *config.Settings, userRepo authrepo.UserRepository) *SettingsService {
	return &SettingsService{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		cfg:          cfg,
		log:          logger.GetLogger("settings"),
	}
}

// GetProfileSettings retrieves a user's profile settings
func (s *SettingsService) GetProfileSettings(ctx context.Context, userId int) (*models.Profile, error) {
	profile, err := s.settingsRepo.GetProfile(ctx, userId)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get profile settings")
		return nil, err
	}

	if profile == nil {
		s.log.Info().Int("user_id", userId).Msg("Profile not found, creating empty profile")
		return &models.Profile{
			UserID:         userId,
			Skills:         []string{},
			WorkExperience: []models.WorkExperience{},
			Education:      []models.Education{},
			Certifications: []models.Certification{},
		}, nil
	}

	return profile, nil
}

// GetWorkExperiences retrieves a user's work experiences
func (s *SettingsService) GetWorkExperiences(ctx context.Context, profileID int) ([]*models.WorkExperience, error) {
	// If the profile ID is 0, it means a profile doesn't exist yet
	if profileID == 0 {
		return []*models.WorkExperience{}, nil
	}

	experiences, err := s.settingsRepo.GetWorkExperiences(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get work experiences")
		return []*models.WorkExperience{}, err
	}

	// Convert to pointer slice
	result := make([]*models.WorkExperience, len(experiences))
	for i := range experiences {
		exp := experiences[i] // Copy to avoid referencing loop variable
		result[i] = &exp
	}
	return result, nil
}

// GetWorkExperienceByID retrieves a single work experience by its ID and verifies it belongs to the given profile
func (s *SettingsService) GetWorkExperienceByID(ctx context.Context, experienceID, profileID int) (*models.WorkExperience, error) {
	experiences, err := s.settingsRepo.GetWorkExperiences(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get work experiences")
		return nil, err
	}

	for _, exp := range experiences {
		if exp.ID == experienceID {
			return &exp, nil
		}
	}

	s.log.Warn().Int("experience_id", experienceID).Int("profile_id", profileID).Msg("Work experience not found or doesn't belong to profile")
	return nil, errors.New("work experience not found")
}

// UpdateWorkExperience updates an existing work experience
func (s *SettingsService) UpdateWorkExperience(ctx context.Context, experience *models.WorkExperience) error {
	updated, err := s.settingsRepo.UpdateWorkExperience(ctx, experience)
	if err != nil {
		s.log.Error().Err(err).Int("experience_id", experience.ID).Msg("Failed to update work experience")
		return err
	}

	// Copy updated fields back
	experience.UpdatedAt = updated.UpdatedAt
	return nil
}

// DeleteWorkExperience deletes a work experience by its ID
// It first verifies the experience belongs to the specified profile
func (s *SettingsService) DeleteWorkExperience(ctx context.Context, experienceID, profileID int) error {
	// Verify the experience belongs to this profile
	_, err := s.GetWorkExperienceByID(ctx, experienceID, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("experience_id", experienceID).Int("profile_id", profileID).
			Msg("Failed to verify work experience before deletion")
		return err
	}

	// Delete the experience
	if err := s.settingsRepo.DeleteWorkExperience(ctx, experienceID); err != nil {
		s.log.Error().Err(err).Int("experience_id", experienceID).Msg("Failed to delete work experience")
		return err
	}

	s.log.Info().Int("experience_id", experienceID).Int("profile_id", profileID).Msg("Successfully deleted work experience")
	return nil
}

func (s *SettingsService) CreateWorkExperience(ctx context.Context, experience *models.WorkExperience) error {
	if err := s.settingsRepo.AddWorkExperience(ctx, experience); err != nil {
		s.log.Error().Err(err).Msg("Failed to add work experience")
		return err
	}
	return nil
}

// GetEducation retrieves a user's education entries
func (s *SettingsService) GetEducation(ctx context.Context, profileID int) ([]*models.Education, error) {
	// If the profile ID is 0, it means a profile doesn't exist yet
	if profileID == 0 {
		return []*models.Education{}, nil
	}

	education, err := s.settingsRepo.GetEducation(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get education entries")
		return []*models.Education{}, err
	}

	// Convert to pointer slice
	result := make([]*models.Education, len(education))
	for i := range education {
		edu := education[i] // Copy to avoid referencing loop variable
		result[i] = &edu
	}
	return result, nil
}

// GetEducationByID retrieves a single education entry by its ID and verifies it belongs to the given profile
func (s *SettingsService) GetEducationByID(ctx context.Context, educationID, profileID int) (*models.Education, error) {
	educationEntries, err := s.settingsRepo.GetEducation(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get education entries")
		return nil, err
	}

	for _, edu := range educationEntries {
		if edu.ID == educationID {
			return &edu, nil
		}
	}

	s.log.Warn().Int("education_id", educationID).Int("profile_id", profileID).Msg("Education entry not found or doesn't belong to profile")
	return nil, errors.New("education entry not found")
}

// CreateEducation adds a new education entry to a user's profile
func (s *SettingsService) CreateEducation(ctx context.Context, education *models.Education) error {
	if err := s.settingsRepo.AddEducation(ctx, education); err != nil {
		s.log.Error().Err(err).Msg("Failed to add education entry")
		return err
	}
	return nil
}

// UpdateEducation updates an existing education entry
func (s *SettingsService) UpdateEducation(ctx context.Context, education *models.Education) error {
	updated, err := s.settingsRepo.UpdateEducation(ctx, education)
	if err != nil {
		s.log.Error().Err(err).Int("education_id", education.ID).Msg("Failed to update education entry")
		return err
	}

	// Copy updated fields back
	education.UpdatedAt = updated.UpdatedAt
	return nil
}

// DeleteEducation deletes an education entry by its ID
// It first verifies the entry belongs to the specified profile
func (s *SettingsService) DeleteEducation(ctx context.Context, educationID, profileID int) error {
	// Verify the education entry belongs to this profile
	_, err := s.GetEducationByID(ctx, educationID, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("education_id", educationID).Int("profile_id", profileID).
			Msg("Failed to verify education entry before deletion")
		return err
	}

	// Delete the education entry
	if err := s.settingsRepo.DeleteEducation(ctx, educationID); err != nil {
		s.log.Error().Err(err).Int("education_id", educationID).Msg("Failed to delete education entry")
		return err
	}

	s.log.Info().Int("education_id", educationID).Int("profile_id", profileID).Msg("Successfully deleted education entry")
	return nil
}

// GetCertifications retrieves a user's certifications
func (s *SettingsService) GetCertifications(ctx context.Context, profileID int) ([]*models.Certification, error) {
	// If the profile ID is 0, it means a profile doesn't exist yet
	if profileID == 0 {
		return []*models.Certification{}, nil
	}

	certifications, err := s.settingsRepo.GetCertifications(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get certifications")
		return []*models.Certification{}, err
	}

	// Convert to pointer slice
	result := make([]*models.Certification, len(certifications))
	for i := range certifications {
		cert := certifications[i] // Copy to avoid referencing loop variable
		result[i] = &cert
	}
	return result, nil
}

// GetCertificationByID retrieves a single certification by its ID and verifies it belongs to the given profile
func (s *SettingsService) GetCertificationByID(ctx context.Context, certificationID, profileID int) (*models.Certification, error) {
	certifications, err := s.settingsRepo.GetCertifications(ctx, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("profile_id", profileID).Msg("Failed to get certifications")
		return nil, err
	}

	for _, cert := range certifications {
		if cert.ID == certificationID {
			return &cert, nil
		}
	}

	s.log.Warn().Int("certification_id", certificationID).Int("profile_id", profileID).Msg("Certification not found or doesn't belong to profile")
	return nil, errors.New("certification not found")
}

// CreateCertification adds a new certification to a user's profile
func (s *SettingsService) CreateCertification(ctx context.Context, certification *models.Certification) error {
	if err := s.settingsRepo.AddCertification(ctx, certification); err != nil {
		s.log.Error().Err(err).Msg("Failed to add certification")
		return err
	}
	return nil
}

// UpdateCertification updates an existing certification
func (s *SettingsService) UpdateCertification(ctx context.Context, certification *models.Certification) error {
	updated, err := s.settingsRepo.UpdateCertification(ctx, certification)
	if err != nil {
		s.log.Error().Err(err).Int("certification_id", certification.ID).Msg("Failed to update certification")
		return err
	}

	// Copy updated fields back
	certification.UpdatedAt = updated.UpdatedAt
	return nil
}

// DeleteCertification deletes a certification by its ID
// It first verifies the certification belongs to the specified profile
func (s *SettingsService) DeleteCertification(ctx context.Context, certificationID, profileID int) error {
	// Verify the certification belongs to this profile
	_, err := s.GetCertificationByID(ctx, certificationID, profileID)
	if err != nil {
		s.log.Error().Err(err).Int("certification_id", certificationID).Int("profile_id", profileID).
			Msg("Failed to verify certification before deletion")
		return err
	}

	// Delete the certification
	if err := s.settingsRepo.DeleteCertification(ctx, certificationID); err != nil {
		s.log.Error().Err(err).Int("certification_id", certificationID).Msg("Failed to delete certification")
		return err
	}

	s.log.Info().Int("certification_id", certificationID).Int("profile_id", profileID).Msg("Successfully deleted certification")
	return nil
}

// UpdateProfile updates a user's profile in the database
func (s *SettingsService) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	if err := s.settingsRepo.UpdateProfile(ctx, profile); err != nil {
		s.log.Error().Err(err).Int("user_id", profile.UserID).Msg("Failed to update profile")
		return err
	}
	return nil
}

// GetSecuritySettings retrieves a user's security settings
func (s *SettingsService) GetSecuritySettings(ctx context.Context, username string) (*models.SecuritySettings, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	activity := models.NewAccountActivity(user.LastLogin, user.CreatedAt)
	return models.NewSecuritySettings(activity), nil
}
