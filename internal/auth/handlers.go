package auth

import (
	"net/http"

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

func (h *AuthHandler) GetChangePasswordPage(c *gin.Context) {
	c.String(200, "Change Password Page")
}

func (h *AuthHandler) GetProfilePage(c *gin.Context) {
	c.String(200, "Profile Page")
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

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	c.String(200, "Change Password")
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	c.String(200, "Profile")
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
