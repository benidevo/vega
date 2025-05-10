package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboardPage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	mockConfig := &config.Settings{
		IsTest: true,
	}

	handler := NewHandler(mockConfig)

	r.GET("/dashboard", func(c *gin.Context) {
		c.Set("username", "testuser")
		handler.GetDashboardPage(c)
	})

	req, _ := http.NewRequest("GET", "/dashboard", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
