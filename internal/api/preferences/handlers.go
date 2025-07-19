package preferences

import (
	"net/http"

	"github.com/benidevo/vega/internal/settings"
	"github.com/gin-gonic/gin"
)

// PreferencesHandler manages API requests for job search preferences
type PreferencesHandler struct {
	settingsService *settings.SettingsService
}

// NewPreferencesHandler creates a new PreferencesHandler instance
func NewPreferencesHandler(settingsService *settings.SettingsService) *PreferencesHandler {
	return &PreferencesHandler{
		settingsService: settingsService,
	}
}

// GetActivePreferences returns active job search preferences for the authenticated user
func (h *PreferencesHandler) GetActivePreferences(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDValue.(int)

	preferences, err := h.settingsService.GetActivePreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve preferences",
		})
		return
	}

	// Transform preferences to API response format
	type preferenceResponse struct {
		ID       string `json:"id"`
		JobTitle string `json:"job_title"`
		Location string `json:"location"`
		MaxAge   int    `json:"max_age"`
	}

	var response []preferenceResponse
	for _, pref := range preferences {
		response = append(response, preferenceResponse{
			ID:       pref.ID,
			JobTitle: pref.JobTitle,
			Location: pref.Location,
			MaxAge:   pref.MaxAge,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": response,
	})
}
