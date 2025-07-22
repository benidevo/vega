package alerts

import (
	"net/http"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
)

// AlertContext defines the rendering context for alerts
type AlertContext string

const (
	ContextGeneral   AlertContext = "general"
	ContextDashboard AlertContext = "dashboard"
	ContextDetail    AlertContext = "detail"
)

// AlertType defines the type of alert
type AlertType string

const (
	TypeError   AlertType = "error"
	TypeSuccess AlertType = "success"
	TypeWarning AlertType = "warning"
	TypeInfo    AlertType = "info"
)

// AlertAction represents an action link in an alert
type AlertAction struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

// AlertData contains the data passed to the alert template
type AlertData struct {
	Type    AlertType     `json:"type"`
	Context AlertContext  `json:"context"`
	Message string        `json:"message"`
	Actions []AlertAction `json:"actions,omitempty"`
}

// RenderError renders an error alert with the appropriate context
// For HTMX requests, it triggers a toast notification
// For 500 errors on non-HTMX requests, it renders a full error page
func RenderError(c *gin.Context, statusCode int, message string, context AlertContext) {
	RenderErrorWithConfig(c, statusCode, message, context, nil)
}

// RenderErrorWithConfig renders an error alert with the appropriate context and optional config
// For HTMX requests, it triggers a toast notification
// For 500 errors on non-HTMX requests, it renders a full error page
func RenderErrorWithConfig(c *gin.Context, statusCode int, message string, context AlertContext, cfg *config.Settings) {
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		TriggerToast(c, message, TypeError)
		c.Status(statusCode)
		return
	}

	// For 500 errors on non-HTMX requests, show full error page
	if statusCode >= 500 {
		templateData := gin.H{
			"title":       "Something Went Wrong",
			"page":        "500",
			"currentYear": time.Now().Year(),
		}

		// Security page is always enabled
		templateData["securityPageEnabled"] = true

		c.HTML(http.StatusInternalServerError, "layouts/base.html", templateData)
		return
	}

	c.Status(statusCode)
}

// RenderSuccess renders a success alert with the appropriate context
// For HTMX requests, it triggers a toast notification
func RenderSuccess(c *gin.Context, message string, context AlertContext, actions ...AlertAction) {
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// For HTMX requests, trigger a toast notification
		TriggerToast(c, message, TypeSuccess)
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusOK)
}
