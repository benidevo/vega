package pages

import (
	"net/http"

	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
)

// Handler manages static page HTTP requests.
type Handler struct {
	cfg      *config.Settings
	renderer *render.HTMLRenderer
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings) *Handler {
	return &Handler{
		cfg:      cfg,
		renderer: render.NewHTMLRenderer(cfg),
	}
}

// GetPrivacyPage renders the privacy policy page.
func (h *Handler) GetPrivacyPage(c *gin.Context) {
	h.renderer.HTML(c, http.StatusOK, "pages/privacy.html", gin.H{
		"title": "Privacy Policy",
	})
}
