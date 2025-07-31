package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRenderer(isCloudMode bool) (*HTMLRenderer, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Settings{
		IsCloudMode: isCloudMode,
	}
	renderer := NewHTMLRenderer(cfg)
	router := gin.New()
	return renderer, router
}

func TestNewHTMLRenderer(t *testing.T) {
	t.Run("should_create_renderer_when_config_provided", func(t *testing.T) {
		cfg := &config.Settings{IsCloudMode: true}
		renderer := NewHTMLRenderer(cfg)

		require.NotNil(t, renderer)
		assert.Equal(t, cfg, renderer.cfg)
	})
}

func TestBaseTemplateData(t *testing.T) {
	t.Run("should_return_base_data_when_no_context_values", func(t *testing.T) {
		renderer, router := setupTestRenderer(false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		router.GET("/", func(ctx *gin.Context) {
			data := renderer.BaseTemplateData(ctx)

			assert.Equal(t, time.Now().Year(), data["currentYear"])
			assert.Equal(t, true, data["securityPageEnabled"])
			assert.Nil(t, data["username"])
			assert.Equal(t, false, data["isCloudMode"])
			assert.Nil(t, data["csrfToken"])
		})

		router.ServeHTTP(w, c.Request)
	})

	t.Run("should_include_context_values_when_set", func(t *testing.T) {
		renderer, router := setupTestRenderer(true)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		router.GET("/", func(ctx *gin.Context) {
			ctx.Set("username", "testuser")
			ctx.Set("csrfToken", "test-csrf-token")

			data := renderer.BaseTemplateData(ctx)

			assert.Equal(t, "testuser", data["username"])
			assert.Equal(t, "test-csrf-token", data["csrfToken"])
			assert.Equal(t, true, data["isCloudMode"])
		})

		router.ServeHTTP(w, c.Request)
	})
}

func TestHTML(t *testing.T) {
	t.Run("should_render_with_merged_data_when_custom_data_provided", func(t *testing.T) {
		renderer, router := setupTestRenderer(false)

		// Mock template for testing
		router.SetHTMLTemplate(mockTemplate())

		w := httptest.NewRecorder()

		router.GET("/", func(ctx *gin.Context) {
			ctx.Set("username", "testuser")

			customData := gin.H{
				"title":   "Test Page",
				"content": "Test Content",
			}

			renderer.HTML(ctx, http.StatusOK, "test.html", customData)
		})

		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "testuser")
		assert.Contains(t, body, "Test Page")
		assert.Contains(t, body, "Test Content")
	})

	t.Run("should_override_base_data_when_custom_data_has_same_keys", func(t *testing.T) {
		renderer, router := setupTestRenderer(false)

		// Mock template for testing
		router.SetHTMLTemplate(mockTemplate())

		w := httptest.NewRecorder()

		router.GET("/", func(ctx *gin.Context) {
			ctx.Set("username", "originaluser")

			customData := gin.H{
				"username": "overriddenuser", // This should override the context value
				"title":    "Test",
			}

			renderer.HTML(ctx, http.StatusOK, "test.html", customData)
		})

		req := httptest.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "overriddenuser")
		assert.NotContains(t, body, "originaluser")
	})
}

func TestError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		title        string
		expectedPage string
	}{
		{
			name:         "should_render_404_page_when_not_found",
			code:         404,
			title:        "Page Not Found",
			expectedPage: "404",
		},
		{
			name:         "should_render_401_page_when_unauthorized",
			code:         401,
			title:        "Unauthorized",
			expectedPage: "401",
		},
		{
			name:         "should_render_403_page_when_forbidden",
			code:         403,
			title:        "Forbidden",
			expectedPage: "403",
		},
		{
			name:         "should_render_500_page_when_server_error",
			code:         500,
			title:        "Internal Server Error",
			expectedPage: "500",
		},
		{
			name:         "should_render_generic_error_page_when_unknown_code",
			code:         418, // I'm a teapot
			title:        "Unknown Error",
			expectedPage: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, router := setupTestRenderer(false)

			// Mock template for testing
			router.SetHTMLTemplate(mockTemplate())

			w := httptest.NewRecorder()

			router.GET("/", func(ctx *gin.Context) {
				renderer.Error(ctx, tt.code, tt.title)
			})

			req := httptest.NewRequest("GET", "/", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.code, w.Code)
			body := w.Body.String()
			assert.Contains(t, body, tt.title)
			assert.Contains(t, body, tt.expectedPage)
		})
	}
}

func TestGetErrorPage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"should_return_404_when_code_is_404", 404, "404"},
		{"should_return_401_when_code_is_401", 401, "401"},
		{"should_return_403_when_code_is_403", 403, "403"},
		{"should_return_500_when_code_is_500", 500, "500"},
		{"should_return_error_when_code_is_unknown", 418, "error"},
		{"should_return_error_when_code_is_200", 200, "error"},
		{"should_return_error_when_code_is_302", 302, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorPage(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a mock template
func mockTemplate() *template.Template {
	tmpl := template.New("test")
	tmpl, _ = tmpl.Parse(`{{.username}} {{.title}} {{.content}} {{.page}}`)
	tmpl.New("test.html").Parse(`{{.username}} {{.title}} {{.content}}`)
	tmpl.New("layouts/base.html").Parse(`{{.title}} {{.page}}`)
	return tmpl
}
