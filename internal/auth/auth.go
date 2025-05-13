package auth

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// SetupAuth initializes and returns an AuthHandler using the provided database connection and configuration settings.
// It sets up the user repository, authentication service, and handler dependencies.
func SetupAuth(db *sql.DB, cfg *config.Settings) *AuthHandler {
	repo := NewSQLiteUserRepository(db)
	service := NewAuthService(repo, cfg)
	handler := NewAuthHandler(service)

	return handler
}

// RegisterPublicRoutes registers public authentication routes (login page and login action)
// to the provided Gin router group using the specified AuthHandler.
func RegisterPublicRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	router.GET("/login", handler.GetLoginPage)
	router.POST("/login", handler.Login)
}

// RegisterPrivateRoutes registers private authentication-related routes to the provided router group.
// It attaches handler functions for endpoints that require authentication, such as logout.
func RegisterPrivateRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	router.POST("/logout", handler.Logout)
}
