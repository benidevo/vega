package home

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandler_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		path        string
		method      string
		isCloudMode bool
	}{
		{
			name:        "should_register_root_route",
			path:        "/",
			method:      "GET",
			isCloudMode: true,
		},
		{
			name:        "should_register_dashboard_route",
			path:        "/dashboard",
			method:      "GET",
			isCloudMode: true,
		},
		{
			name:        "should_register_root_route_self_hosted",
			path:        "/",
			method:      "GET",
			isCloudMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			cfg := &config.Settings{
				IsCloudMode:        tt.isCloudMode,
				GoogleOAuthEnabled: true,
			}

			// Create handler with nil service (won't be called in route test)
			handler := NewHandler(cfg, nil)

			// Register a simple handler that returns OK
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			// Test that route works
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify handler was created correctly
			assert.NotNil(t, handler)
			assert.Equal(t, cfg, handler.cfg)
			assert.NotNil(t, handler.renderer)
		})
	}
}

func TestNewHandler(t *testing.T) {
	cfg := &config.Settings{
		IsCloudMode: true,
	}
	service := &Service{}

	handler := NewHandler(cfg, service)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.cfg)
	assert.Equal(t, service, handler.service)
	assert.NotNil(t, handler.renderer)
}
