package preferences

import (
	"net/http"

	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/quota"
	"github.com/benidevo/vega/internal/settings"
	"github.com/gin-gonic/gin"
)

// PreferencesHandler manages API requests for job search preferences
type PreferencesHandler struct {
	settingsService *settings.SettingsService
	quotaService    *quota.UnifiedService
	log             *logger.PrivacyLogger
}

// NewPreferencesHandler creates a new PreferencesHandler instance
func NewPreferencesHandler(settingsService *settings.SettingsService, quotaService *quota.UnifiedService) *PreferencesHandler {
	return &PreferencesHandler{
		settingsService: settingsService,
		quotaService:    quotaService,
		log:             logger.GetPrivacyLogger("preferences_api"),
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

	// Check search quota before returning preferences
	canSearch, err := h.quotaService.CheckQuota(c.Request.Context(), userID, quota.QuotaTypeSearchRuns, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check quota",
		})
		return
	}

	if !canSearch.Allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":        "Search quota exceeded",
			"quota_status": canSearch.Status,
		})
		return
	}

	preferences, err := h.settingsService.GetActivePreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve preferences",
		})
		return
	}

	quotaStatus, err := h.quotaService.GetAllQuotaStatus(c.Request.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).
			Int("user_id", userID).
			Msg("Failed to get quota status for preferences response")
		quotaStatus = nil
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

	result := gin.H{
		"preferences": response,
	}

	if quotaStatus != nil {
		result["quota_status"] = quotaStatus
	}

	c.JSON(http.StatusOK, result)
}

// RecordJobSearchResults records job search results from the extension
func (h *PreferencesHandler) RecordJobSearchResults(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDValue.(int)

	var payload struct {
		PreferenceID string `json:"preference_id"`
		JobsFound    int    `json:"jobs_found"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate payload
	if payload.PreferenceID == "" || payload.JobsFound < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid preference_id or jobs_found count",
		})
		return
	}
	err := h.quotaService.RecordUsage(c.Request.Context(), userID, quota.QuotaTypeJobSearch, map[string]interface{}{
		"count": payload.JobsFound,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record usage",
		})
		return
	}

	err = h.quotaService.RecordUsage(c.Request.Context(), userID, quota.QuotaTypeSearchRuns, nil)
	if err != nil {
		h.log.Error().Err(err).
			Int("user_id", userID).
			Str("preference_id", payload.PreferenceID).
			Msg("Failed to record search run usage")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usage recorded successfully",
	})
}
