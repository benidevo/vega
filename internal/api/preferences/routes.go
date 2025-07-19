package preferences

import "github.com/gin-gonic/gin"

// RegisterRoutes registers preference-related API routes
func RegisterRoutes(apiGroup *gin.RouterGroup, handler *PreferencesHandler) {
	preferencesGroup := apiGroup.Group("/preferences")
	{
		preferencesGroup.GET("/active", handler.GetActivePreferences)
	}
}
