package home

import (
	"net/http"

	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
)

// Handler manages home page related HTTP requests.
type Handler struct {
	cfg      *config.Settings
	service  *Service
	renderer *render.HTMLRenderer
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings, service *Service) *Handler {
	return &Handler{
		cfg:      cfg,
		service:  service,
		renderer: render.NewHTMLRenderer(cfg),
	}
}

// GetHomePage renders the home page template with dynamic user data.
func (h *Handler) GetHomePage(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		// Fallback to static template if no authentication
		emptyHomeData := NewHomePageData(0, "")

		h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
			"title":              emptyHomeData.Title,
			"page":               emptyHomeData.Page,
			"googleOAuthEnabled": h.cfg.GoogleOAuthEnabled,
			"isCloudMode":        h.cfg.IsCloudMode,
			"showOnboarding":     emptyHomeData.ShowOnboarding,
			"stats":              emptyHomeData.Stats,
			"recentJobs":         emptyHomeData.RecentJobs,
			"hasJobs":            emptyHomeData.HasJobs,
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

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
		"title":          homeData.Title,
		"page":           homeData.Page,
		"activeNav":      "home",
		"pageTitle":      "Dashboard",
		"stats":          homeData.Stats,
		"recentJobs":     homeData.RecentJobs,
		"hasJobs":        homeData.HasJobs,
		"showOnboarding": homeData.ShowOnboarding,
		"quotaStatus":    homeData.QuotaStatus,
	})
}
