package settings

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benidevo/vega/internal/common/alerts"
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
}

// CRUDService defines the interface for CRUD operations
type CRUDService interface {
	GetProfileSettings(ctx context.Context, userID int) (*models.Profile, error)
	CreateEntity(ctx *gin.Context, entity CRUDEntity) error
	UpdateEntity(ctx *gin.Context, entity CRUDEntity) error
	DeleteEntity(ctx *gin.Context, entityID, profileID int, entityType string) error
	GetEntityByID(ctx *gin.Context, entityID, profileID int, entityType string) (CRUDEntity, error)
}

// NewBaseSettingsHandler creates a new base settings handler
func NewBaseSettingsHandler(service CRUDService, metadata EntityMetadata) *BaseSettingsHandler {
	return &BaseSettingsHandler{
		service:  service,
		metadata: metadata,
	}
}

// GetAddPage handles the request to display the add entity page
func (h *BaseSettingsHandler) GetAddPage(c *gin.Context) {
	username, _ := c.Get("username")
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID.(int))
	if err != nil {
		profile = &models.Profile{}
	}

	templateData := gin.H{
		"title":          fmt.Sprintf("Add %s", h.metadata.Name),
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"profile":        profile,
		fmt.Sprintf("isAdding%s", h.metadata.Name): true,
	}

	c.HTML(http.StatusOK, "layouts/base.html", templateData)
}

// GetEditPage handles the request to display the edit entity page
func (h *BaseSettingsHandler) GetEditPage(c *gin.Context) {
	username, _ := c.Get("username")
	userID := c.GetInt("userID")
	entityIDStr := c.Param("id")

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "layouts/base.html", gin.H{
			"title":       "Bad Request",
			"page":        "404",
			"currentYear": time.Now().Year(),
		})
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":       "Something Went Wrong",
			"page":        "500",
			"currentYear": time.Now().Year(),
		})
		return
	}

	entity, err := h.service.GetEntityByID(c, entityID, profile.ID, h.metadata.Name)
	if err != nil {
		c.HTML(http.StatusNotFound, "layouts/base.html", gin.H{
			"title":       "Not Found",
			"page":        "404",
			"currentYear": time.Now().Year(),
		})
		return
	}

	templateData := gin.H{
		"title":          fmt.Sprintf("Edit %s", h.metadata.Name),
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"profile":        profile,
		fmt.Sprintf("isEditing%s", h.metadata.Name): true,
		strings.ToLower(h.metadata.Name):            entity,
	}

	c.HTML(http.StatusOK, "layouts/base.html", templateData)
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
		alerts.RenderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create %s: %s", h.metadata.Name, err.Error()), alerts.ContextDashboard)
		return
	}

	c.Header("HX-Trigger", "closeModal")

	// Set a URL parameter to trigger scrolling after page load
	var scrollSection string
	switch strings.ToLower(h.metadata.Name) {
	case "experience":
		scrollSection = "experience"
	case "education":
		scrollSection = "education"
	case "certification":
		scrollSection = "certification"
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
		alerts.RenderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update %s: %s", h.metadata.Name, err.Error()), alerts.ContextDashboard)
		return
	}

	// Show success message on the edit page
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "general",
		"message": fmt.Sprintf("%s updated successfully", h.metadata.Name),
	})
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
