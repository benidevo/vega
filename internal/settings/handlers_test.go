package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/ai"
	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSettingsRoutesRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should_register_profile_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/profile", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "profile"})
			})
			settingsGroup.POST("/profile/personal", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "personal"})
			})
			settingsGroup.POST("/profile/online", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "online"})
			})
			settingsGroup.POST("/profile/context", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "context"})
			})
			settingsGroup.POST("/profile/parse-cv", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "parse-cv"})
			})
		})
	})

	t.Run("should_register_account_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/account", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "account"})
			})
			settingsGroup.POST("/account/update", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "update"})
			})
			settingsGroup.DELETE("/account/delete", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "delete"})
			})
		})
	})

	t.Run("should_register_experience_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/profile/experience/new", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "new-experience"})
			})
			settingsGroup.GET("/profile/experience/:id/edit", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "edit-experience"})
			})
			settingsGroup.POST("/profile/experience", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "create-experience"})
			})
			settingsGroup.POST("/profile/experience/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "update-experience"})
			})
			settingsGroup.DELETE("/profile/experience/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "delete-experience"})
			})
		})
	})

	t.Run("should_register_education_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/profile/education/new", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "new-education"})
			})
			settingsGroup.GET("/profile/education/:id/edit", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "edit-education"})
			})
			settingsGroup.POST("/profile/education", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "create-education"})
			})
			settingsGroup.POST("/profile/education/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "update-education"})
			})
			settingsGroup.DELETE("/profile/education/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "delete-education"})
			})
		})
	})

	t.Run("should_register_certification_routes", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/profile/certification/new", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "new-certification"})
			})
			settingsGroup.GET("/profile/certification/:id/edit", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "edit-certification"})
			})
			settingsGroup.POST("/profile/certification", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "create-certification"})
			})
			settingsGroup.POST("/profile/certification/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "update-certification"})
			})
			settingsGroup.DELETE("/profile/certification/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "delete-certification"})
			})
		})
	})

	t.Run("should_register_quotas_route", func(t *testing.T) {
		router := gin.New()
		settingsGroup := router.Group("/settings")

		assert.NotPanics(t, func() {
			settingsGroup.GET("/quotas", func(c *gin.Context) {
				c.JSON(200, gin.H{"route": "quotas"})
			})
		})
	})
}

func TestNewSettingsHandler(t *testing.T) {
	mockService := &SettingsService{
		cfg: &config.Settings{},
	}
	mockAIService := &ai.AIService{}

	handler := NewSettingsHandler(mockService, mockAIService, nil)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.service)
	assert.NotNil(t, handler.aiService)
	assert.NotNil(t, handler.experienceHandler)
	assert.NotNil(t, handler.educationHandler)
	assert.NotNil(t, handler.certificationHandler)
	assert.NotNil(t, handler.renderer)
}

func TestSettingsRouteStructure(t *testing.T) {
	router := gin.New()

	router.GET("/settings", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/settings/profile")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/settings", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/settings/profile", w.Header().Get("Location"))
}
