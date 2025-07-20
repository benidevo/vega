package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	csrfTokenLength  = 32
	csrfCookieName   = "csrf_token"
	csrfHeaderName   = "X-CSRF-Token"
	csrfCookieMaxAge = 86400 // 24 hours in seconds
)

func generateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for API routes (they use Bearer tokens)
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// Skip CSRF for OAuth routes (they have state parameter)
		if strings.HasPrefix(c.Request.URL.Path, "/auth/google") || strings.HasPrefix(c.Request.URL.Path, "/auth/callback") {
			c.Next()
			return
		}

		// Skip CSRF for static assets
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Next()
			return
		}

		// Get or generate CSRF token
		cookie, err := c.Cookie(csrfCookieName)
		var token string

		if err != nil || cookie == "" {
			token, err = generateCSRFToken()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.SetCookie(
				csrfCookieName,
				token,
				csrfCookieMaxAge,
				"/",
				"",
				c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https"), // Secure in production (HTTPS or via reverse proxy)
				true, // HttpOnly
			)
		} else {
			token = cookie
		}

		// Add token to gin context for templates
		c.Set("csrfToken", token)

		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			// Check header first (for HTMX requests)
			headerToken := c.GetHeader(csrfHeaderName)

			// Check form value if header not present
			if headerToken == "" {
				headerToken = c.PostForm("csrf_token")
			}

			if headerToken != token {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF Token"})
				return
			}
		}

		c.Next()
	}
}
