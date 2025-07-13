package auth

import (
	"github.com/gin-gonic/gin"
)

// GetUserID safely retrieves the user ID from the Gin context.
// It returns the user ID and a boolean indicating whether it was found.
// This helper ensures consistent error handling across the application.
func GetUserID(c *gin.Context) (int, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	userID, ok := userIDValue.(int)
	if !ok {
		return 0, false
	}

	return userID, true
}

// MustGetUserID retrieves the user ID from the Gin context.
// It panics if the user ID is not found or is not an integer.
// This should only be used in handlers that are guaranteed to run
// after authentication middleware.
func MustGetUserID(c *gin.Context) int {
	userID, ok := GetUserID(c)
	if !ok {
		panic("userID not found in context or invalid type")
	}
	return userID
}
