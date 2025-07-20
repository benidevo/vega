package home

import (
	"net/http"

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
	_, exists := c.Get("userID")
	if !exists {
		// In cloud mode, show landing page to non-authenticated users
		if h.cfg.IsCloudMode {
			h.renderer.HTML(c, http.StatusOK, "landing/index.html", gin.H{
				"title": "Vega AI - AI-Powered Job Search Assistant",
			})
			return
		}

		// In self-hosted mode, redirect to login
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
		return
	}

	// If authenticated, redirect to jobs dashboard
	c.Redirect(http.StatusTemporaryRedirect, "/jobs")
}
