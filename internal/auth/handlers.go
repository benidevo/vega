package auth

import (
	"fmt"
	"net/http"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates and returns a new AuthHandler with the provided AuthService.
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// GetLoginPage renders the login page template.
func (h *AuthHandler) GetLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title": "Login",
		"page":  "login",
	})
}

// Login handles user authentication by validating credentials from the POST form.
// On success, it sets a session cookie and redirects to the dashboard.
// On failure, it returns an unauthorized status and error message for HTMX response swapping.
func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	token, loginErr := h.service.Login(c.Request.Context(), username, password)
	if loginErr != nil {
		c.Header("HX-Reswap", "innerHTML")
		c.Header("HX-Retarget", "#form-response")

		c.String(http.StatusUnauthorized, loginErr.Error())
		return
	}

	c.SetCookie("token", token, 3600, "/", "", false, true)
	c.Header("HX-Redirect", "/jobs")
	c.Status(http.StatusOK)
}

// Logout logs out the current user by clearing the authentication cookie and redirecting to the home page.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// AuthMiddleware is a Gin middleware that checks for a valid JWT token in the "token" cookie.
// If the token is valid, it sets user information (userID, username, role) in the context.
// If the token is missing or invalid, it redirects the user to the login page.
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
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
	service *GoogleAuthService
	config  *config.Settings
}

// NewGoogleAuthHandler creates and returns a new GoogleAuthHandler with the provided service
func NewGoogleAuthHandler(service *GoogleAuthService, cfg *config.Settings) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		service: service,
		config:  cfg,
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

	token, err := h.service.Authenticate(c.Request.Context(), code)
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

	c.SetCookie("token", token, 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/jobs")
}
