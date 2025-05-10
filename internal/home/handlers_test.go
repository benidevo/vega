package home

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetHomePage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockConfig := &config.Settings{
		IsTest: true,
	}

	handler := NewHandler(mockConfig)

	r.GET("/", handler.GetHomePage)

	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
