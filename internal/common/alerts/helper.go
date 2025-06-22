package alerts

import (
	"net/http"
	"time"

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
// For 500 errors on non-HTMX requests, it renders a full error page
func RenderError(c *gin.Context, statusCode int, message string, context AlertContext) {
	// For 500 errors on non-HTMX requests, always show full error page
	if statusCode >= 500 && c.GetHeader("HX-Request") != "true" {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":       "Something Went Wrong",
			"page":        "500",
			"currentYear": time.Now().Year(),
		})
		return
	}

	// For all other cases (HTMX requests or non-500 errors), return a partial alert
	c.HTML(statusCode, "partials/alert.html", gin.H{
		"type":    string(TypeError),
		"context": string(context),
		"message": message,
	})
}

// RenderSuccess renders a success alert with the appropriate context
func RenderSuccess(c *gin.Context, message string, context AlertContext, actions ...AlertAction) {
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    string(TypeSuccess),
		"context": string(context),
		"message": message,
		"actions": actions,
	})
}

// RenderWarning renders a warning alert with the appropriate context
func RenderWarning(c *gin.Context, message string, context AlertContext) {
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    string(TypeWarning),
		"context": string(context),
		"message": message,
	})
}

// RenderInfo renders an info alert with the appropriate context
func RenderInfo(c *gin.Context, message string, context AlertContext) {
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    string(TypeInfo),
		"context": string(context),
		"message": message,
	})
}
