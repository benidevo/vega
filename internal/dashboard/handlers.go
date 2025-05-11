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
	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Dashboard",
		"page":        "dashboard",
		"activeNav":   "jobs",
		"pageTitle":   "Job Matches",
		"currentYear": time.Now().Year(),
		"username":    username,
	})
}
