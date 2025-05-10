package home

import (
	"github.com/benidevo/prospector/internal/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler manages home page related HTTP requests.
type Handler struct {
	cfg *config.Settings
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

// GetHomePage renders the home page template.
func (h *Handler) GetHomePage(c *gin.Context) {
	if h.cfg != nil && h.cfg.IsTest {
		// In test mode, just return a status code
		c.Status(http.StatusOK)
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Home",
		"currentYear": time.Now().Year(),
		"page":        "home",
	})
}
