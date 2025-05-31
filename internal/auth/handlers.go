package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/benidevo/ascentio/internal/auth/models"
	"github.com/benidevo/ascentio/internal/auth/services"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service *services.AuthService
	cfg     *config.Settings
}

// NewAuthHandler creates and returns a new AuthHandler with the provided AuthService.
func NewAuthHandler(service *services.AuthService, cfg *config.Settings) *AuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,
	}
}

// GetLoginPage renders the login page template.
func (h *AuthHandler) GetLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title": "Login",
		"page":  "login",
	})
}

// Login handles user authentication by validating credentials from the request.
// On success, it sets access and refresh token cookies and redirects to the dashboard.
// On failure, it returns an unauthorized status and error message for HTMX response swapping.
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		h.service.LogError(err)
		c.HTML(http.StatusUnauthorized, "partials/form-error.html", gin.H{
			"message": models.ErrInvalidCredentials.Error(),
		})
		return
	}

	accessToken, refreshToken, loginErr := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if loginErr != nil {
		h.service.LogError(loginErr)
		c.HTML(http.StatusUnauthorized, "partials/form-error.html", gin.H{
			"message": loginErr.Error(),
		})

		return
	}

	sameSite := parseSameSiteMode(h.cfg.CookieSameSite)

	// Set access token cookie (shorter-lived)
	c.SetSameSite(sameSite)
	c.SetCookie(
		"token",
		accessToken,
		int(h.cfg.AccessTokenExpiry.Seconds()),
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)

	// Set refresh token cookie (longer-lived)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(h.cfg.RefreshTokenExpiry.Seconds()),
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)

	c.Header("HX-Redirect", "/jobs")
	c.Status(http.StatusOK)
}

// Logout logs out the current user by clearing the authentication cookie and redirecting to the home page.
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear both access and refresh token cookies
	sameSite := parseSameSiteMode(h.cfg.CookieSameSite)

	c.SetSameSite(sameSite)
	c.SetCookie("token", "", -1, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, true)

	c.Redirect(http.StatusFound, "/")
}

// RefreshToken validates a refresh token and issues a new access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}

	accessToken, err := h.service.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	sameSite := parseSameSiteMode(h.cfg.CookieSameSite)

	// Set the new access token
	c.SetSameSite(sameSite)
	c.SetCookie(
		"token",
		accessToken,
		int(h.cfg.AccessTokenExpiry.Seconds()),
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AuthMiddleware is a Gin middleware that checks for a valid JWT token in the "token" cookie.
// If the token is valid, it sets user information (userID, username, role) in the context.
// If the token is missing or invalid, it attempts to use refresh token, or redirects to login.
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")

		// If access token is missing or invalid, try to refresh it
		if err != nil || tokenString == "" {
			refreshToken, refreshErr := c.Cookie("refresh_token")
			if refreshErr == nil && refreshToken != "" {
				newAccessToken, refreshErr := h.service.RefreshAccessToken(c.Request.Context(), refreshToken)
				if refreshErr == nil {
					sameSite := parseSameSiteMode(h.cfg.CookieSameSite)
					c.SetSameSite(sameSite)
					c.SetCookie(
						"token",
						newAccessToken,
						int(h.cfg.AccessTokenExpiry.Seconds()),
						"/",
						h.cfg.CookieDomain,
						h.cfg.CookieSecure,
						true,
					)

					tokenString = newAccessToken
				}
			}

			// If still no valid token, redirect to login
			if tokenString == "" {
				c.Redirect(http.StatusFound, "/auth/login")
				c.Abort()
				return
			}
		}

		claims, err := h.service.VerifyToken(tokenString)
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GoogleAuthHandler handles authentication requests using Google OAuth.
type GoogleAuthHandler struct {
	service *services.GoogleAuthService
	cfg     *config.Settings
}

// NewGoogleAuthHandler creates and returns a new GoogleAuthHandler with the provided service
func NewGoogleAuthHandler(service *services.GoogleAuthService, cfg *config.Settings) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		service: service,
		cfg:     cfg,
	}
}

// HandleLogin initiates the Google OAuth login flow by redirecting the user to the authentication URL.
func (h *GoogleAuthHandler) HandleLogin(c *gin.Context) {
	authURL := h.service.GetAuthURL()
	c.Redirect(http.StatusFound, authURL)
}

// HandleCallback processes the OAuth2 callback from Google, exchanges the code for a token,
// sets the authentication cookie, and redirects the user to the jobs page.
// For errors, redirects to login page with appropriate error messages.
func (h *GoogleAuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		h.service.LogError(fmt.Errorf("missing code in callback"))
		// Redirect to login page with error message
		c.HTML(http.StatusBadRequest, "layouts/base.html", gin.H{
			"title": "Login",
			"page":  "login",
			"error": "Invalid authentication request. Please try again.",
		})
		return
	}

	accessToken, refreshToken, err := h.service.Authenticate(c.Request.Context(), code, "")
	if err != nil {
		h.service.LogError(err)

		errorMessage := "Authentication failed. Please try again later."

		c.HTML(http.StatusOK, "layouts/base.html", gin.H{
			"title": "Login",
			"page":  "login",
			"error": errorMessage,
		})
		return
	}

	sameSite := parseSameSiteMode(h.cfg.CookieSameSite)

	c.SetSameSite(sameSite)
	c.SetCookie(
		"token",
		accessToken,
		int(h.cfg.AccessTokenExpiry.Seconds()),
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)

	// Set refresh token cookie
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(h.cfg.RefreshTokenExpiry.Seconds()),
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)

	c.Redirect(http.StatusFound, "/jobs")
}

// Helper function to parse SameSite mode from string
func parseSameSiteMode(sameSiteStr string) http.SameSite {
	switch strings.ToLower(sameSiteStr) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
