package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Authentication endpoints (stricter)
	AuthRate string // e.g., "5-M" = 5 per minute

	// AI endpoints (expensive operations)
	AIRate string // e.g., "10-5M" = 10 per 5 minutes

	// General API endpoints
	GeneralRate string // e.g., "100-M" = 100 per minute
}

// DefaultRateLimitConfig returns production-ready defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AuthRate:    "5-M",   // 5 login attempts per minute
		AIRate:      "10-5M", // 10 AI operations per 5 minutes
		GeneralRate: "100-M", // 100 requests per minute general
	}
}

// PathRateLimiter creates a rate limiting middleware with different limits for different paths
func PathRateLimiter() gin.HandlerFunc {
	return PathRateLimiterWithConfig(DefaultRateLimitConfig())
}

// PathRateLimiterWithConfig creates a rate limiting middleware with custom configuration
func PathRateLimiterWithConfig(config RateLimitConfig) gin.HandlerFunc {
	// Create memory store (you can swap this for Redis in production)
	store := memory.NewStore()

	// Create rate limiters for different endpoint types
	authLimiter := createLimiter(config.AuthRate, store)
	aiLimiter := createLimiter(config.AIRate, store)
	generalLimiter := createLimiter(config.GeneralRate, store)

	// Create middleware instances
	authMiddleware := mgin.NewMiddleware(authLimiter)
	aiMiddleware := mgin.NewMiddleware(aiLimiter)
	generalMiddleware := mgin.NewMiddleware(generalLimiter)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Apply different rate limits based on path patterns
		switch {
		// Authentication endpoints
		case strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/register"):
			authMiddleware(c)

		// AI-powered endpoints (expensive operations)
		case strings.Contains(path, "/analyze") || strings.Contains(path, "/cover-letter"):
			aiMiddleware(c)

		// Health check and metrics - no rate limiting
		case path == "/health" || path == "/metrics":
			c.Next()

		// Everything else gets general rate limiting
		default:
			generalMiddleware(c)
		}
	}
}

// createLimiter creates a limiter instance with the given rate string
func createLimiter(rateStr string, store limiter.Store) *limiter.Limiter {
	rate, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		// Fall back to a safe default if parsing fails
		rate = limiter.Rate{
			Period: time.Minute,
			Limit:  100,
		}
	}

	return limiter.New(store, rate)
}
