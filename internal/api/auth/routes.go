package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes registers API authentication-related routes to the provided Gin router group.
func RegisterRoutes(router *gin.RouterGroup, handler *AuthAPIHandler) {
	router.POST("/google", handler.ExchangeTokenForJWT)
	router.POST("/refresh", handler.RefreshToken)
	router.POST("/login", handler.Login)
}
