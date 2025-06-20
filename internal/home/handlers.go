package home

import (
	"net/http"

	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
)

// Handler manages home page related HTTP requests.
type Handler struct {
	cfg     *config.Settings
	service *Service
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings, service *Service) *Handler {
	return &Handler{
		cfg:     cfg,
		service: service,
	}
}

// GetHomePage renders the home page template with dynamic user data.
func (h *Handler) GetHomePage(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		// Fallback to static template if no authentication
		emptyHomeData := NewHomePageData(0, "")

		c.HTML(http.StatusOK, "layouts/base.html", gin.H{
			"title":          emptyHomeData.Title,
			"page":           emptyHomeData.Page,
			"showOnboarding": emptyHomeData.ShowOnboarding,
			"stats":          emptyHomeData.Stats,
			"recentJobs":     emptyHomeData.RecentJobs,
			"hasJobs":        emptyHomeData.HasJobs,
		})
		return
	}

	userID := userIDValue.(int)
	username, _ := c.Get("username")
	usernameStr := ""
	if username != nil {
		if str, ok := username.(string); ok {
			usernameStr = str
		}
	}

	homeData, err := h.service.GetHomePageData(c.Request.Context(), userID, usernameStr)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load homepage data", alerts.ContextGeneral)
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":     homeData.Title,
		"page":      homeData.Page,
		"username":  homeData.Username,
		"activeNav": "home",
		"pageTitle": "Dashboard",

		"stats":          homeData.Stats,
		"recentJobs":     homeData.RecentJobs,
		"hasJobs":        homeData.HasJobs,
		"showOnboarding": homeData.ShowOnboarding,
	})
}
