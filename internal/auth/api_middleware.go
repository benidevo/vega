package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// APIAuthMiddleware creates a middleware for API authentication using Bearer tokens
func (h *AuthHandler) APIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Debug().Msg("No Authorization header found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Debug().Msg("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication failed",
			})
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := h.service.VerifyToken(token)
		if err != nil {
			log.Debug().Err(err).Msg("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication failed",
			})
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			log.Debug().Msg("Token is not an access token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication failed",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
