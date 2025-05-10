package auth

import (
	"github.com/benidevo/prospector/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service *AuthService
	log     zerolog.Logger
}

// NewAuthHandler creates and returns a new AuthHandler with the provided AuthService and a logger instance.
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     logger.GetLogger("auth"),
	}
}

// GetLoginPage renders the login page template.
func (h *AuthHandler) GetLoginPage(c *gin.Context) {
	c.HTML(200, "auth/login.html", gin.H{
		"title": "Login",
	})
}

func (h *AuthHandler) GetChangePasswordPage(c *gin.Context) {
	c.String(200, "Change Password Page")
}

func (h *AuthHandler) GetProfilePage(c *gin.Context) {
	c.String(200, "Profile Page")
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.String(200, "Login")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.String(200, "Logout")
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	c.String(200, "Change Password")
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	c.String(200, "Profile")
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
