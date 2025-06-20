package settings

import (
	"context"
	"fmt"

	authrepo "github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/settings/interfaces"
	"github.com/benidevo/ascentio/internal/settings/models"
	"github.com/benidevo/ascentio/internal/settings/services"
	"github.com/gin-gonic/gin"
)

// SettingsService provides business logic for user settings management
type SettingsService struct {
	userRepo         authrepo.UserRepository
	settingsRepo     interfaces.SettingsRepository
	cfg              *config.Settings
	log              *logger.PrivacyLogger
	centralValidator *services.CentralizedValidator
}

// NewSettingsService creates a new SettingsService instance
func NewSettingsService(settingsRepo interfaces.SettingsRepository, cfg *config.Settings, userRepo authrepo.UserRepository) *SettingsService {
	return &SettingsService{
		userRepo:         userRepo,
		settingsRepo:     settingsRepo,
		cfg:              cfg,
		log:              logger.GetPrivacyLogger("settings"),
		centralValidator: services.NewCentralizedValidator(),
	}
}

// GetProfileSettings retrieves a user's profile settings
func (s *SettingsService) GetProfileSettings(ctx context.Context, userID int) (*models.Profile, error) {
	profile, err := s.settingsRepo.GetProfile(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "profile_get_failed").
			Str("user_ref", fmt.Sprintf("user_%d", userID)).
			Msg("Failed to get profile settings")
		return nil, models.WrapError(models.ErrFailedToGetSettings, err)
	}
	if profile != nil {
		return profile, nil
	}

	s.log.Info().
		Str("event", "profile_empty_created").
		Str("user_ref", fmt.Sprintf("user_%d", userID)).
		Msg("Profile not found, creating empty profile")
	return &models.Profile{
		UserID:         userID,
		Skills:         []string{},
		WorkExperience: []models.WorkExperience{},
		Education:      []models.Education{},
		Certifications: []models.Certification{},
	}, nil
}

// GetProfileWithRelated retrieves a user's profile with all related entities
func (s *SettingsService) GetProfileWithRelated(ctx context.Context, userID int) (*models.Profile, error) {
	return s.settingsRepo.GetProfileWithRelated(ctx, userID)
}

// UpdateProfile updates a user's profile with centralized validation
func (s *SettingsService) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	profile.Sanitize()

	if err := s.centralValidator.ValidateProfile(profile); err != nil {
		s.log.Error().Err(err).Msg("Profile validation failed")
		return err
	}

	if err := s.settingsRepo.UpdateProfile(ctx, profile); err != nil {
		s.log.Error().Err(err).Int("user_id", profile.UserID).Msg("Failed to update profile")
		return models.WrapError(models.ErrFailedToUpdateSettings, err)
	}
	return nil
}

// CreateEntity creates a new entity (Experience, Education, or Certification)
func (s *SettingsService) CreateEntity(ctx *gin.Context, entity CRUDEntity) error {
	entity.Sanitize()

	if err := s.validateEntity(entity); err != nil {
		s.log.Error().Err(err).Msg("Entity validation failed")
		return err
	}

	switch e := entity.(type) {
	case *models.WorkExperience:
		if err := s.settingsRepo.AddWorkExperience(ctx.Request.Context(), e); err != nil {
			s.log.Error().Err(err).Int("profile_id", e.ProfileID).Msg("Failed to add work experience")
			return models.WrapError(models.ErrFailedToCreateWorkExperience, err)
		}
		return nil
	case *models.Education:
		if err := s.settingsRepo.AddEducation(ctx.Request.Context(), e); err != nil {
			s.log.Error().Err(err).Int("profile_id", e.ProfileID).Msg("Failed to add education entry")
			return models.WrapError(models.ErrFailedToCreateEducation, err)
		}
		return nil
	case *models.Certification:
		if err := s.settingsRepo.AddCertification(ctx.Request.Context(), e); err != nil {
			s.log.Error().Err(err).Int("profile_id", e.ProfileID).Msg("Failed to add certification")
			return models.WrapError(models.ErrFailedToCreateCertification, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported entity type")
	}
}

// UpdateEntity updates an existing entity with validation
func (s *SettingsService) UpdateEntity(ctx *gin.Context, entity CRUDEntity) error {
	entity.Sanitize()

	if err := s.validateEntity(entity); err != nil {
		s.log.Error().Err(err).Msg("Entity validation failed")
		return err
	}

	switch e := entity.(type) {
	case *models.WorkExperience:
		_, err := s.settingsRepo.UpdateWorkExperience(ctx.Request.Context(), e)
		if err != nil {
			s.log.Error().Err(err).Int("experience_id", e.ID).Msg("Failed to update work experience")
			if err == models.ErrWorkExperienceNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToUpdateWorkExperience, err)
		}
		return nil
	case *models.Education:
		_, err := s.settingsRepo.UpdateEducation(ctx.Request.Context(), e)
		if err != nil {
			s.log.Error().Err(err).Int("education_id", e.ID).Msg("Failed to update education entry")
			if err == models.ErrEducationNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToUpdateEducation, err)
		}
		return nil
	case *models.Certification:
		_, err := s.settingsRepo.UpdateCertification(ctx.Request.Context(), e)
		if err != nil {
			s.log.Error().Err(err).Int("certification_id", e.ID).Msg("Failed to update certification")
			if err == models.ErrCertificationNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToUpdateCertification, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported entity type")
	}
}

// DeleteEntity deletes an entity with ownership verification
func (s *SettingsService) DeleteEntity(ctx *gin.Context, entityID, profileID int, entityType string) error {
	_, err := s.GetEntityByID(ctx, entityID, profileID, entityType)
	if err != nil {
		return err
	}

	switch entityType {
	case "Experience":
		if err := s.settingsRepo.DeleteWorkExperience(ctx.Request.Context(), entityID); err != nil {
			s.log.Error().Err(err).Int("experience_id", entityID).Msg("Failed to delete work experience")
			if err == models.ErrWorkExperienceNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToDeleteWorkExperience, err)
		}
		s.log.Info().Int("experience_id", entityID).Int("profile_id", profileID).Msg("Successfully deleted work experience")
		return nil
	case "Education":
		if err := s.settingsRepo.DeleteEducation(ctx.Request.Context(), entityID); err != nil {
			s.log.Error().Err(err).Int("education_id", entityID).Msg("Failed to delete education entry")
			if err == models.ErrEducationNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToDeleteEducation, err)
		}
		s.log.Info().Int("education_id", entityID).Int("profile_id", profileID).Msg("Successfully deleted education entry")
		return nil
	case "Certification":
		if err := s.settingsRepo.DeleteCertification(ctx.Request.Context(), entityID); err != nil {
			s.log.Error().Err(err).Int("certification_id", entityID).Msg("Failed to delete certification")
			if err == models.ErrCertificationNotFound {
				return err
			}
			return models.WrapError(models.ErrFailedToDeleteCertification, err)
		}
		s.log.Info().Int("certification_id", entityID).Int("profile_id", profileID).Msg("Successfully deleted certification")
		return nil
	default:
		return fmt.Errorf("unsupported entity type")
	}
}

// GetEntityByID retrieves an entity by ID with ownership verification
func (s *SettingsService) GetEntityByID(ctx *gin.Context, entityID, profileID int, entityType string) (CRUDEntity, error) {
	entity, err := s.settingsRepo.GetEntityByID(ctx.Request.Context(), entityID, profileID, entityType)
	if err != nil {
		return nil, err
	}
	// Convert to CRUDEntity
	if crudEntity, ok := entity.(CRUDEntity); ok {
		return crudEntity, nil
	}
	return nil, fmt.Errorf("entity does not implement CRUDEntity interface")
}

// ValidateContext validates context with word count limit
func (s *SettingsService) ValidateContext(context string) error {
	return s.centralValidator.ValidateWordCount(context, 1000, "Context")
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

// validateEntity validates an entity based on its type
func (s *SettingsService) validateEntity(entity CRUDEntity) error {
	switch e := entity.(type) {
	case *models.WorkExperience:
		return s.centralValidator.ValidateWorkExperience(e)
	case *models.Education:
		return s.centralValidator.ValidateEducation(e)
	case *models.Certification:
		return s.centralValidator.ValidateCertification(e)
	default:
		return fmt.Errorf("unsupported entity type for validation")
	}
}
