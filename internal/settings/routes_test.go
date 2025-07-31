package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should_redirect_root_to_profile", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")
		handler := &SettingsHandler{}

		RegisterRoutes(settingsGroup, handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/settings", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/settings/profile", w.Header().Get("Location"))
	})

	t.Run("should_register_expected_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")
		handler := &SettingsHandler{}

		RegisterRoutes(settingsGroup, handler)

		// Check that routes are registered by looking at the router's routes
		routes := router.Routes()

		expectedPaths := []string{
			"/settings",
			"/settings/profile",
			"/settings/account",
			"/settings/quotas",
		}

		for _, expectedPath := range expectedPaths {
			found := false
			for _, route := range routes {
				if route.Path == expectedPath {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected route %s to be registered", expectedPath)
		}
	})
}
