package auth

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes registers public authentication routes (login page and login action)
// to the provided Gin router group using the specified AuthHandler.
func RegisterPublicRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	router.GET("/login", handler.GetLoginPage)
	router.POST("/login", handler.Login)
	router.POST("/refresh", handler.RefreshToken)
}

// RegisterPrivateRoutes registers private authentication-related routes to the provided router group.
// It attaches handler functions for endpoints that require authentication, such as logout.
func RegisterPrivateRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	router.POST("/logout", handler.Logout)
}

// RegisterGoogleAuthRoutes registers Google authentication routes to the provided router group.
// It attaches handler functions for Google login and callback endpoints.
func RegisterGoogleAuthRoutes(router *gin.RouterGroup, handler *GoogleAuthHandler) {
	router.GET("/google/login", handler.HandleLogin)
	router.GET("/google/callback", handler.HandleCallback)
}
