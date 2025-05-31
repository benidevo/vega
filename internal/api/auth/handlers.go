package auth

import (
	"fmt"
	"net/http"

	"github.com/benidevo/ascentio/internal/auth/services"
	"github.com/gin-gonic/gin"
)

// AuthAPIHandler handles authentication-related API requests using the provided OAuth service and authentication service.
type AuthAPIHandler struct {
	oauthService *services.GoogleAuthService
	authService  *services.AuthService
}

// NewAuthAPIHandler creates and returns a new AuthAPIHandler with the provided GoogleAuthService and Authentication service.
func NewAuthAPIHandler(oathService *services.GoogleAuthService, authService *services.AuthService) *AuthAPIHandler {
	return &AuthAPIHandler{
		oauthService: oathService,
		authService:  authService,
	}
}

// ExchangeTokenForJWT handles the exchange of an OAuth authorization code for a JWT access token.
//
// It expects a JSON body with "code" and "redirect_uri" fields, and returns the JWT token on success.
func (h *AuthAPIHandler) ExchangeTokenForJWT(ctx *gin.Context) {
	var request struct {
		Code        string `json:"code" binding:"required"`
		RedirectURI string `json:"redirect_uri" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.oauthService.LogError(fmt.Errorf("failed to bind request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	h.oauthService.LogError(fmt.Errorf("received token: %s : %s", request.Code, request.RedirectURI))

	accessToken, refreshToken, err := h.oauthService.Authenticate(ctx.Request.Context(), request.Code, request.RedirectURI)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "failed to exchange code for JWT"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": accessToken, "refresh_token": refreshToken})
}

// RefreshToken handles the token refresh request by validating the provided refresh token
// and returning a new access token if successful.
func (h *AuthAPIHandler) RefreshToken(ctx *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.oauthService.LogError(fmt.Errorf("failed to bind request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(ctx.Request.Context(), request.RefreshToken)
	if err != nil {
		h.oauthService.LogError(fmt.Errorf("failed to refresh access token: %v", err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "failed to refresh access token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": accessToken})
}
