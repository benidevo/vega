package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "same-origin")

		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' https://unpkg.com https://cdn.jsdelivr.net; " +
			"font-src 'self' data:; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'"
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}
