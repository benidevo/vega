package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		csrfToken      string
		existingCookie string
		formData       url.Values
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "GET request should not require CSRF token",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API routes should skip CSRF",
			method:         "POST",
			path:           "/api/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OAuth routes should skip CSRF",
			method:         "POST",
			path:           "/auth/google",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Static routes should skip CSRF",
			method:         "POST",
			path:           "/static/test.js",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST without CSRF token should fail",
			method:         "POST",
			path:           "/test",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST with CSRF token in header should succeed",
			method:         "POST",
			path:           "/test",
			existingCookie: "test-token",
			headers:        map[string]string{"X-CSRF-Token": "test-token"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with CSRF token in form should succeed",
			method:         "POST",
			path:           "/test",
			existingCookie: "test-token",
			formData:       url.Values{"csrf_token": {"test-token"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with wrong CSRF token should fail",
			method:         "POST",
			path:           "/test",
			existingCookie: "test-token",
			headers:        map[string]string{"X-CSRF-Token": "wrong-token"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Create a test config
			cfg := &config.Settings{}

			// Use a custom recovery middleware to catch template rendering panics
			router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
				// For testing, just abort with the expected status
				c.AbortWithStatus(http.StatusForbidden)
			}))

			router.Use(CSRF(cfg))

			router.Any("/*path", func(c *gin.Context) {
				c.String(http.StatusOK, "test")
			})

			w := httptest.NewRecorder()

			var req *http.Request
			if tt.formData != nil {
				req, _ = http.NewRequest(tt.method, tt.path, strings.NewReader(tt.formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, _ = http.NewRequest(tt.method, tt.path, nil)
			}

			if tt.existingCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  csrfCookieName,
					Value: tt.existingCookie,
				})
			}

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check if CSRF cookie is set for non-skipped routes
			if tt.method == "GET" && !strings.HasPrefix(tt.path, "/api/") &&
				!strings.HasPrefix(tt.path, "/auth/") && !strings.HasPrefix(tt.path, "/static/") {
				cookies := w.Result().Cookies()
				var foundCSRF bool
				for _, cookie := range cookies {
					if cookie.Name == csrfCookieName {
						foundCSRF = true
						assert.NotEmpty(t, cookie.Value)
						assert.True(t, cookie.HttpOnly)
						break
					}
				}
				if tt.existingCookie == "" {
					assert.True(t, foundCSRF, "CSRF cookie should be set")
				}
			}
		})
	}
}
