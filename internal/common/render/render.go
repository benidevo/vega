package render

import (
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
)

// HTMLRenderer provides common rendering functionality for all handlers
type HTMLRenderer struct {
	cfg *config.Settings
}

// NewHTMLRenderer creates a new HTMLRenderer instance
func NewHTMLRenderer(cfg *config.Settings) *HTMLRenderer {
	return &HTMLRenderer{cfg: cfg}
}

// BaseTemplateData returns common template data that all handlers need
func (r *HTMLRenderer) BaseTemplateData(c *gin.Context) gin.H {
	username, _ := c.Get("username")
	return gin.H{
		"currentYear":         time.Now().Year(),
		"securityPageEnabled": true, // Always enabled
		"username":            username,
		"isCloudMode":         r.cfg.IsCloudMode,
	}
}

// HTML renders an HTML template with merged base and custom data
func (r *HTMLRenderer) HTML(c *gin.Context, code int, template string, data gin.H) {
	// Start with base data
	finalData := r.BaseTemplateData(c)

	// Merge custom data
	for key, value := range data {
		finalData[key] = value
	}

	c.HTML(code, template, finalData)
}

// Error renders an error page with the appropriate status code
func (r *HTMLRenderer) Error(c *gin.Context, code int, title string) {
	r.HTML(c, code, "layouts/base.html", gin.H{
		"title": title,
		"page":  getErrorPage(code),
	})
}

// getErrorPage returns the appropriate error page template name based on status code
func getErrorPage(code int) string {
	switch code {
	case 404:
		return "404"
	case 401:
		return "401"
	case 403:
		return "403"
	case 500:
		return "500"
	default:
		return "error"
	}
}
