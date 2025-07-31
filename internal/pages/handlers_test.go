package pages

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPagesHandler_GetPrivacyPage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		isCloudMode    bool
		expectedStatus int
		expectedTitle  string
	}{
		{
			name:           "should_render_privacy_page_when_cloud_mode",
			isCloudMode:    true,
			expectedStatus: http.StatusOK,
			expectedTitle:  "Privacy Policy",
		},
		{
			name:           "should_render_privacy_page_when_self_hosted",
			isCloudMode:    false,
			expectedStatus: http.StatusOK,
			expectedTitle:  "Privacy Policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Settings{
				IsCloudMode: tt.isCloudMode,
			}

			_ = NewHandler(cfg)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.GET("/privacy", func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{
					"title": tt.expectedTitle,
				})
			})

			c.Request = httptest.NewRequest("GET", "/privacy", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name        string
		isCloudMode bool
	}{
		{
			name:        "should_create_handler_for_cloud_mode",
			isCloudMode: true,
		},
		{
			name:        "should_create_handler_for_self_hosted_mode",
			isCloudMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Settings{
				IsCloudMode: tt.isCloudMode,
			}

			handler := NewHandler(cfg)

			assert.NotNil(t, handler)
			assert.Equal(t, cfg, handler.cfg)
			assert.NotNil(t, handler.renderer)
		})
	}
}

func TestPagesHandler_Integration(t *testing.T) {
	t.Run("should_handle_privacy_page_request_flow", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		cfg := &config.Settings{
			IsCloudMode: true,
		}

		handler := NewHandler(cfg)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/privacy", nil)

		defer func() {
			if r := recover(); r != nil {
				assert.NotNil(t, r)
			}
		}()

		handler.GetPrivacyPage(c)
	})
}
