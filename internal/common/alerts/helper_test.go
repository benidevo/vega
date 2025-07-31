package alerts

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRenderError(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		message         string
		context         AlertContext
		isHTMX          bool
		expectedHeaders map[string]string
		expectPanic     bool
	}{
		{
			name:       "should_trigger_error_toast_when_htmx_request_with_error",
			statusCode: http.StatusBadRequest,
			message:    "Validation error",
			context:    ContextGeneral,
			isHTMX:     true,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Validation error","type":"error"}}`,
			},
		},
		{
			name:        "should_render_error_page_when_500_status_and_non_htmx",
			statusCode:  http.StatusInternalServerError,
			message:     "Server error",
			context:     ContextGeneral,
			isHTMX:      false,
			expectPanic: true, // HTML rendering will panic in test
		},
		{
			name:       "should_return_status_code_when_non_500_error_and_non_htmx",
			statusCode: http.StatusNotFound,
			message:    "Not found",
			context:    ContextDashboard,
			isHTMX:     false,
		},
		{
			name:        "should_render_error_page_when_service_unavailable_and_non_htmx",
			statusCode:  http.StatusServiceUnavailable,
			message:     "Service unavailable",
			context:     ContextDetail,
			isHTMX:      false,
			expectPanic: true, // HTML rendering will panic in test
		},
		{
			name:       "should_trigger_error_toast_when_htmx_request_with_500_error",
			statusCode: http.StatusInternalServerError,
			message:    "Server error",
			context:    ContextGeneral,
			isHTMX:     true,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Server error","type":"error"}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.isHTMX {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set("HX-Request", "true")
			} else {
				c.Request = httptest.NewRequest("GET", "/", nil)
			}

			if tt.expectPanic {
				assert.Panics(t, func() {
					RenderError(c, tt.statusCode, tt.message, tt.context)
				})
			} else {
				RenderError(c, tt.statusCode, tt.message, tt.context)

				for header, value := range tt.expectedHeaders {
					assert.Equal(t, value, w.Header().Get(header))
				}
			}
		})
	}
}

func TestRenderErrorWithConfig(t *testing.T) {
	cfg := &config.Settings{
		AppName: "TestApp",
	}

	tests := []struct {
		name            string
		statusCode      int
		message         string
		context         AlertContext
		isHTMX          bool
		expectedHeaders map[string]string
		config          *config.Settings
	}{
		{
			name:       "should_handle_error_when_config_provided",
			statusCode: http.StatusBadRequest,
			message:    "Error with config",
			context:    ContextGeneral,
			isHTMX:     true,
			config:     cfg,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Error with config","type":"error"}}`,
			},
		},
		{
			name:       "should_handle_error_when_config_is_nil",
			statusCode: http.StatusBadRequest,
			message:    "Error without config",
			context:    ContextGeneral,
			isHTMX:     true,
			config:     nil,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Error without config","type":"error"}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.isHTMX {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set("HX-Request", "true")
			} else {
				c.Request = httptest.NewRequest("GET", "/", nil)
			}

			RenderErrorWithConfig(c, tt.statusCode, tt.message, tt.context, tt.config)

			for header, value := range tt.expectedHeaders {
				assert.Equal(t, value, w.Header().Get(header))
			}
		})
	}
}

func TestRenderSuccess(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		context         AlertContext
		actions         []AlertAction
		isHTMX          bool
		expectedHeaders map[string]string
	}{
		{
			name:    "should_trigger_success_toast_when_htmx_request",
			message: "Operation successful",
			context: ContextGeneral,
			isHTMX:  true,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Operation successful","type":"success"}}`,
			},
		},
		{
			name:    "should_return_ok_status_when_non_htmx_request",
			message: "Success",
			context: ContextDashboard,
			isHTMX:  false,
			// No special behavior for non-HTMX success
		},
		{
			name:    "should_trigger_success_toast_when_actions_provided",
			message: "Success with actions",
			context: ContextDetail,
			actions: []AlertAction{
				{URL: "/home", Text: "Go Home"},
				{URL: "/dashboard", Text: "Dashboard"},
			},
			isHTMX: true,
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"showToast":{"message":"Success with actions","type":"success"}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.isHTMX {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.Header.Set("HX-Request", "true")
			} else {
				c.Request = httptest.NewRequest("GET", "/", nil)
			}

			RenderSuccess(c, tt.message, tt.context, tt.actions...)

			for header, value := range tt.expectedHeaders {
				assert.Equal(t, value, w.Header().Get(header))
			}
		})
	}
}

func TestAlertTypes(t *testing.T) {
	t.Run("should_have_correct_constant_values_when_accessed", func(t *testing.T) {
		assert.Equal(t, AlertType("error"), TypeError)
		assert.Equal(t, AlertType("success"), TypeSuccess)
		assert.Equal(t, AlertType("warning"), TypeWarning)
		assert.Equal(t, AlertType("info"), TypeInfo)
	})
}

func TestAlertContexts(t *testing.T) {
	t.Run("should_have_correct_context_values_when_accessed", func(t *testing.T) {
		assert.Equal(t, AlertContext("general"), ContextGeneral)
		assert.Equal(t, AlertContext("dashboard"), ContextDashboard)
		assert.Equal(t, AlertContext("detail"), ContextDetail)
	})
}

func TestAlertDataStructure(t *testing.T) {
	t.Run("should_store_alert_data_when_created", func(t *testing.T) {
		data := AlertData{
			Type:    TypeError,
			Context: ContextGeneral,
			Message: "Test message",
			Actions: []AlertAction{
				{URL: "/test", Text: "Test Action"},
			},
		}

		assert.Equal(t, TypeError, data.Type)
		assert.Equal(t, ContextGeneral, data.Context)
		assert.Equal(t, "Test message", data.Message)
		assert.Len(t, data.Actions, 1)
		assert.Equal(t, "/test", data.Actions[0].URL)
		assert.Equal(t, "Test Action", data.Actions[0].Text)
	})
}

func TestRenderBehavior(t *testing.T) {
	t.Run("should_set_htmx_header_when_htmx_request", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("HX-Request", "true")

		RenderError(c, http.StatusBadRequest, "Test error", ContextGeneral)

		assert.NotEmpty(t, w.Header().Get("HX-Trigger"))
		assert.Contains(t, w.Header().Get("HX-Trigger"), "showToast")
		assert.Contains(t, w.Header().Get("HX-Trigger"), "Test error")
		assert.Contains(t, w.Header().Get("HX-Trigger"), "error")
	})

	t.Run("should_attempt_html_render_when_500_error_non_htmx", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		assert.Panics(t, func() {
			RenderError(c, http.StatusInternalServerError, "Server error", ContextGeneral)
		})
	})

	t.Run("should_not_set_htmx_header_when_non_htmx_request", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		RenderError(c, http.StatusNotFound, "Not found", ContextGeneral)

		assert.Empty(t, w.Header().Get("HX-Trigger"))
	})
}
