package preferences

import (
	"github.com/benidevo/vega/internal/settings"
	"github.com/gin-gonic/gin"
)

// Setup initializes the preferences API module
func Setup(apiGroup *gin.RouterGroup, settingsService *settings.SettingsService) {
	handler := NewPreferencesHandler(settingsService)
	RegisterRoutes(apiGroup, handler)
}
