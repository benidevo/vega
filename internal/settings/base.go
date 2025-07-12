package settings

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/settings/models"
	"github.com/gin-gonic/gin"
)

// CRUDEntity defines the interface for entities that can be handled by the base handler
type CRUDEntity interface {
	GetID() int
	GetProfileID() int
	Validate() error
	Sanitize()
}

// FormBindable defines the interface for entities that can be bound from form data
type FormBindable interface {
	BindFromForm(c *gin.Context) error
}

// EntityMetadata holds metadata about an entity type
type EntityMetadata struct {
	Name       string
	URLPrefix  string
	CreateFunc func() CRUDEntity
}

// BaseSettingsHandler provides generic CRUD operations for profile entities
type BaseSettingsHandler struct {
	service  CRUDService
	metadata EntityMetadata
	renderer *render.HTMLRenderer
}

// CRUDService defines the interface for CRUD operations
type CRUDService interface {
	GetProfileSettings(ctx context.Context, userID int) (*models.Profile, error)
	CreateEntity(ctx *gin.Context, entity CRUDEntity) error
	UpdateEntity(ctx *gin.Context, entity CRUDEntity) error
	DeleteEntity(ctx *gin.Context, entityID, profileID int, entityType string) error
	GetEntityByID(ctx *gin.Context, entityID, profileID int, entityType string) (CRUDEntity, error)
	GetConfig() *config.Settings
}

// NewBaseSettingsHandler creates a new base settings handler
func NewBaseSettingsHandler(service CRUDService, metadata EntityMetadata) *BaseSettingsHandler {
	return &BaseSettingsHandler{
		service:  service,
		metadata: metadata,
		renderer: render.NewHTMLRenderer(service.GetConfig()),
	}
}

// GetAddPage handles the request to display the add entity page
func (h *BaseSettingsHandler) GetAddPage(c *gin.Context) {
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID.(int))
	if err != nil {
		profile = &models.Profile{}
	}

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
		"title":          fmt.Sprintf("Add %s", h.metadata.Name),
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"profile":        profile,
		fmt.Sprintf("isAdding%s", h.metadata.Name): true,
	})
}

// GetEditPage handles the request to display the edit entity page
func (h *BaseSettingsHandler) GetEditPage(c *gin.Context) {
	userID := c.GetInt("userID")
	entityIDStr := c.Param("id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		h.renderer.Error(c, http.StatusBadRequest, "Bad Request")
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		h.renderer.Error(c, http.StatusInternalServerError, "Something Went Wrong")
		return
	}

	entity, err := h.service.GetEntityByID(c, entityID, profile.ID, h.metadata.Name)
	if err != nil {
		h.renderer.Error(c, http.StatusNotFound, "Not Found")
		return
	}

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
		"title":          fmt.Sprintf("Edit %s", h.metadata.Name),
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"profile":        profile,
		fmt.Sprintf("isEditing%s", h.metadata.Name): true,
		strings.ToLower(h.metadata.Name):            entity,
	})
}

// HandleCreate processes the creation of a new entity
func (h *BaseSettingsHandler) HandleCreate(c *gin.Context) {
	userID := c.GetInt("userID")

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
		return
	}

	entity := h.metadata.CreateFunc()
	if bindable, ok := entity.(FormBindable); ok {
		if err := bindable.BindFromForm(c); err != nil {
			alerts.RenderError(c, http.StatusBadRequest, err.Error(), alerts.ContextDashboard)
			return
		}
	}

	// Set the profile ID
	if we, ok := entity.(*models.WorkExperience); ok {
		we.ProfileID = profile.ID
	} else if ed, ok := entity.(*models.Education); ok {
		ed.ProfileID = profile.ID
	} else if cert, ok := entity.(*models.Certification); ok {
		cert.ProfileID = profile.ID
	}

	if err := h.service.CreateEntity(c, entity); err != nil {
		statusCode, message := h.getErrorDetails(err)

		// For non-HTMX requests, render the form page with the error
		if c.GetHeader("HX-Request") != "true" {
			username, _ := c.Get("username")
			templateData := gin.H{
				"title":               fmt.Sprintf("Add %s", h.metadata.Name),
				"page":                "settings-profile",
				"activeNav":           "settings",
				"activeSettings":      "profile",
				"pageTitle":           "Profile Settings",
				"currentYear":         time.Now().Year(),
				"securityPageEnabled": h.service.GetConfig().SecurityPageEnabled,
				"username":            username,
				"profile":             profile,
				fmt.Sprintf("isAdding%s", h.metadata.Name): true,
				"error":                          message,
				strings.ToLower(h.metadata.Name): entity,
			}

			c.HTML(statusCode, "layouts/base.html", templateData)
			return
		}

		// For HTMX requests, render the error alert
		alerts.RenderError(c, statusCode, message, alerts.ContextDashboard)
		return
	}

	c.Header("HX-Trigger", "closeModal")

	// Set a URL parameter to trigger scrolling after page load
	var scrollSection string
	normalizedName := strings.ToLower(h.metadata.Name)
	switch models.EntityType(normalizedName) {
	case models.EntityTypeExperience:
		scrollSection = models.EntityTypeExperience.Lower()
	case models.EntityTypeEducation:
		scrollSection = models.EntityTypeEducation.Lower()
	case models.EntityTypeCertification:
		scrollSection = models.EntityTypeCertification.Lower()
	default:
		scrollSection = "profile" // Fallback to profile section for unmatched cases
	}

	c.Redirect(http.StatusSeeOther, "/settings/profile?scroll="+scrollSection)
}

// HandleUpdate processes the update of an existing entity
func (h *BaseSettingsHandler) HandleUpdate(c *gin.Context) {
	userID := c.GetInt("userID")
	entityIDStr := c.Param("id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, fmt.Sprintf("Invalid %s ID format", h.metadata.Name), alerts.ContextDashboard)
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
		return
	}

	entity, err := h.service.GetEntityByID(c, entityID, profile.ID, h.metadata.Name)
	if err != nil {
		alerts.RenderError(c, http.StatusNotFound, fmt.Sprintf("%s not found or you don't have permission to edit it", h.metadata.Name), alerts.ContextDashboard)
		return
	}

	if bindable, ok := entity.(FormBindable); ok {
		if err := bindable.BindFromForm(c); err != nil {
			alerts.RenderError(c, http.StatusBadRequest, err.Error(), alerts.ContextDashboard)
			return
		}
	}

	if err := h.service.UpdateEntity(c, entity); err != nil {
		statusCode, message := h.getErrorDetails(err)
		alerts.RenderError(c, statusCode, message, alerts.ContextDashboard)
		return
	}

	// Show success message on the edit page
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "general",
		"message": fmt.Sprintf("%s updated successfully", h.metadata.Name),
	})
}

// getErrorDetails determines the appropriate status code and message for an error
func (h *BaseSettingsHandler) getErrorDetails(err error) (int, string) {
	if err == nil {
		return http.StatusInternalServerError, "Unknown error occurred"
	}

	var unwrapped error
	if errors.Is(err, models.ErrFailedToCreateWorkExperience) ||
		errors.Is(err, models.ErrFailedToUpdateWorkExperience) ||
		errors.Is(err, models.ErrFailedToCreateEducation) ||
		errors.Is(err, models.ErrFailedToUpdateEducation) ||
		errors.Is(err, models.ErrFailedToCreateCertification) ||
		errors.Is(err, models.ErrFailedToUpdateCertification) {
		unwrapped = errors.Unwrap(err)
		if unwrapped != nil {
			// Use the unwrapped error for checking
			err = unwrapped
		}
	}

	// Get the error message
	errMsg := err.Error()

	// List of validation error patterns
	validationPatterns := []string{
		"cannot be in the future",
		"must be after",
		"must not exceed",
		"is required",
		"must be positive",
		"cannot be empty",
		"must be a valid",
		"invalid",
		"contains invalid characters",
		"field is required",
		"expiry date must be after issue date",
		"end date must be empty when position is current",
	}

	// Check if the error message contains any validation patterns
	for _, pattern := range validationPatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			// For validation errors, return just the validation message
			// If it was wrapped, use the unwrapped error message
			if unwrapped != nil {
				return http.StatusBadRequest, unwrapped.Error()
			}
			return http.StatusBadRequest, errMsg
		}
	}

	// For all other errors, return 500 with a generic message
	return http.StatusInternalServerError, fmt.Sprintf("Failed to process %s", h.metadata.Name)
}

// HandleDelete processes the deletion of an entity
func (h *BaseSettingsHandler) HandleDelete(c *gin.Context) {
	userID := c.GetInt("userID")
	entityIDStr := c.Param("id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		alerts.RenderError(c, http.StatusBadRequest, fmt.Sprintf("Invalid %s ID format", h.metadata.Name), alerts.ContextDashboard)
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
		return
	}

	if err := h.service.DeleteEntity(c, entityID, profile.ID, h.metadata.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete %s: %s", h.metadata.Name, err.Error()),
		})
		return
	}

	c.Header("HX-Redirect", "/settings/profile")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s deleted successfully", h.metadata.Name)})
}
