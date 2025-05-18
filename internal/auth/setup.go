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

// SetupGoogleAuth initializes and returns a GoogleAuthHandler using the provided configuration settings.
// It sets up the GoogleAuthService and handler dependencies.
func SetupGoogleAuth(cfg *config.Settings, db *sql.DB) (*GoogleAuthHandler, error) {
	repo := NewSQLiteUserRepository(db)
	service, err := NewGoogleAuthService(cfg, repo)
	if err != nil {
		return nil, err
	}
	handler := NewGoogleAuthHandler(service, cfg)

	return handler, nil
}

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
