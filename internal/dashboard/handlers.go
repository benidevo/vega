package dashboard

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// Handler manages dashboard related HTTP requests.
type Handler struct {
	cfg *config.Settings
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

// GetDashboardPage renders the dashboard page template.
func (h *Handler) GetDashboardPage(c *gin.Context) {
	if h.cfg != nil && h.cfg.IsTest {
		// In test mode, just return a status code
		c.Status(http.StatusOK)
		return
	}

	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Dashboard",
		"page":        "dashboard",
		"currentYear": time.Now().Year(),
		"username":    username,
	})
}
