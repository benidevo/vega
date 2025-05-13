package settings

import (
	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// Setup creates a new settings handler and returns it without registering routes
func Setup(cfg *config.Settings) *SettingsHandler {
	service := NewSettingsService(nil, cfg) // I will implement the service later
	return NewSettingsHandler(service, cfg)
}

// RegisterRoutes registers settings-related routes on the given router group
func RegisterRoutes(settingsGroup *gin.RouterGroup, handler *SettingsHandler) {
	// Main settings page
	settingsGroup.GET("", handler.GetSettingsHomePage)

	// Profile settings
	settingsGroup.GET("/profile", handler.GetProfileSettingsPage)

	// Security settings
	settingsGroup.GET("/security", handler.GetSecuritySettingsPage)

	// Notification settings
	settingsGroup.GET("/notifications", handler.GetNotificationSettingsPage)

	// Experience form routes
	settingsGroup.GET("/profile/experience/form", handler.GetExperienceForm)
	settingsGroup.GET("/profile/experience/:id/edit", handler.GetExperienceEditForm)
	settingsGroup.POST("/profile/experience", handler.CreateExperienceForm)

	// Education form routes
	settingsGroup.GET("/profile/education/form", handler.GetEducationForm)
	settingsGroup.GET("/profile/education/:id/edit", handler.GetEducationEditForm)
	settingsGroup.POST("/profile/education", handler.CreateEducationForm)

	// Certification form routes
	settingsGroup.GET("/profile/certification/form", handler.GetCertificationForm)
	settingsGroup.GET("/profile/certification/:id/edit", handler.GetCertificationEditForm)
	settingsGroup.POST("/profile/certification", handler.CreateCertificationForm)

	// Award form routes
	settingsGroup.GET("/profile/award/form", handler.GetAwardForm)
	settingsGroup.GET("/profile/award/:id/edit", handler.GetAwardEditForm)
	settingsGroup.POST("/profile/award", handler.CreateAwardForm)
}
