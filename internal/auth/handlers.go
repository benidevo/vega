package auth

import (
	"github.com/benidevo/prospector/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type AuthHandler struct {
	service *AuthService
	log     zerolog.Logger
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     logger.GetLogger("auth"),
	}
}

func (h *AuthHandler) GetLoginPage(c *gin.Context) {
	c.String(200, "Login Page")
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
