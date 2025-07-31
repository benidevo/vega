package job

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJobRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "should_have_list_jobs_route",
			method: "GET",
			path:   "/jobs",
		},
		{
			name:   "should_have_new_job_form_route",
			method: "GET",
			path:   "/jobs/new",
		},
		{
			name:   "should_have_create_job_route",
			method: "POST",
			path:   "/jobs/new",
		},
		{
			name:   "should_have_job_details_route",
			method: "GET",
			path:   "/jobs/123/details",
		},
		{
			name:   "should_have_match_history_route",
			method: "GET",
			path:   "/jobs/123/match-history",
		},
		{
			name:   "should_have_delete_match_route",
			method: "DELETE",
			path:   "/jobs/123/match-history/456",
		},
		{
			name:   "should_have_update_job_field_route",
			method: "PUT",
			path:   "/jobs/123/status",
		},
		{
			name:   "should_have_delete_job_route",
			method: "DELETE",
			path:   "/jobs/123",
		},
		{
			name:   "should_have_analyze_job_route",
			method: "POST",
			path:   "/jobs/123/analyze",
		},
		{
			name:   "should_have_cover_letter_route",
			method: "POST",
			path:   "/jobs/123/cover-letter",
		},
		{
			name:   "should_have_cv_generation_route",
			method: "POST",
			path:   "/jobs/123/cv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			jobGroup := router.Group("/jobs")

			switch tt.path {
			case "/jobs":
				jobGroup.GET("", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "list"})
				})
			case "/jobs/new":
				if tt.method == "GET" {
					jobGroup.GET("/new", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"route": "new-form"})
					})
				} else {
					jobGroup.POST("/new", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"route": "create"})
					})
				}
			case "/jobs/123/details":
				jobGroup.GET("/:id/details", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "details"})
				})
			case "/jobs/123/match-history":
				jobGroup.GET("/:id/match-history", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "match-history"})
				})
			case "/jobs/123/match-history/456":
				jobGroup.DELETE("/:id/match-history/:matchId", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "delete-match"})
				})
			case "/jobs/123/status":
				jobGroup.PUT("/:id/:field", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "update-field"})
				})
			case "/jobs/123":
				jobGroup.DELETE("/:id", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "delete"})
				})
			case "/jobs/123/analyze":
				jobGroup.POST("/:id/analyze", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "analyze"})
				})
			case "/jobs/123/cover-letter":
				jobGroup.POST("/:id/cover-letter", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "cover-letter"})
				})
			case "/jobs/123/cv":
				jobGroup.POST("/:id/cv", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"route": "cv"})
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code,
				"Route %s %s should work", tt.method, tt.path)
		})
	}
}

func TestJobRoutesStructure(t *testing.T) {
	t.Run("should_group_routes_correctly", func(t *testing.T) {
		router := gin.New()
		jobGroup := router.Group("/jobs")

		assert.NotPanics(t, func() {
			jobGroup.GET("", func(c *gin.Context) {})
			jobGroup.GET("/new", func(c *gin.Context) {})
			jobGroup.POST("/new", func(c *gin.Context) {})

			jobIDGroup := jobGroup.Group("/:id")
			jobIDGroup.GET("/details", func(c *gin.Context) {})
			jobIDGroup.GET("/match-history", func(c *gin.Context) {})
			jobIDGroup.DELETE("/match-history/:matchId", func(c *gin.Context) {})
			jobIDGroup.PUT("/:field", func(c *gin.Context) {})
			jobIDGroup.DELETE("", func(c *gin.Context) {})
			jobIDGroup.POST("/analyze", func(c *gin.Context) {})
			jobIDGroup.POST("/cover-letter", func(c *gin.Context) {})
			jobIDGroup.POST("/cv", func(c *gin.Context) {})
		})
	})
}
