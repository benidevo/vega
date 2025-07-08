package middleware

import (
	"net/http"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/storage"
	"github.com/gin-gonic/gin"
)

const (
	// UserStorageKey is the context key for user storage
	UserStorageKey = "user_storage"
)

// TenantIsolation middleware ensures each user has their own storage instance
func TenantIsolation(factory *storage.Factory, cfg *config.Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tenant isolation if not in cloud mode
		if !cfg.IsCloudMode || !cfg.MultiTenancyEnabled {
			c.Next()
			return
		}

		// Get user from context set by auth middleware
		username, exists := c.Get("username")
		if !exists {
			// For unauthenticated requests, continue without storage
			c.Next()
			return
		}

		userEmail, ok := username.(string)
		if !ok || userEmail == "" {
			c.Next()
			return
		}

		userStorage, err := factory.GetUserStorage(c.Request.Context(), userEmail)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
				"title":       "Something Went Wrong",
				"page":        "500",
				"currentYear": time.Now().Year(),
			})
			c.Abort()
			return
		}

		c.Set(UserStorageKey, userStorage)

		c.Next()
	}
}

// GetUserStorage retrieves the user storage from context
func GetUserStorage(c *gin.Context) (storage.UserStorage, bool) {
	value, exists := c.Get(UserStorageKey)
	if !exists {
		return nil, false
	}

	userStorage, ok := value.(storage.UserStorage)
	return userStorage, ok
}

// RequireUserStorage ensures that user storage is available
func RequireUserStorage() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetUserStorage(c)
		if !exists {
			c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
				"title":       "Something Went Wrong",
				"page":        "500",
				"currentYear": time.Now().Year(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
