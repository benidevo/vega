package auth

import (
	"context"
	"net/http"

	"github.com/benidevo/vega/internal/common/logger"
	"github.com/gin-gonic/gin"
)

// oauthService defines OAuth operations needed by the API handler
type oauthService interface {
	Authenticate(ctx context.Context, code, redirectURI string) (accessToken, refreshToken string, err error)
}

// authService defines auth operations needed by the API handler
type authService interface {
	Login(ctx context.Context, username, password string) (string, string, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
}

// AuthAPIHandler handles authentication-related API requests using the provided OAuth service and authentication service.
type AuthAPIHandler struct {
	oauthService oauthService
	authService  authService
	log          *logger.PrivacyLogger
}

// NewAuthAPIHandler creates and returns a new AuthAPIHandler with the provided GoogleAuthService and Authentication service.
func NewAuthAPIHandler(oauthService oauthService, authService authService) *AuthAPIHandler {
	return &AuthAPIHandler{
		oauthService: oauthService,
		authService:  authService,
		log:          logger.GetPrivacyLogger("api_auth"),
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
		h.log.Error().Err(err).Msg("Failed to bind OAuth token exchange request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	h.log.Debug().
		Str("redirect_uri", request.RedirectURI).
		Msg("OAuth token exchange request received")

	accessToken, refreshToken, err := h.oauthService.Authenticate(ctx.Request.Context(), request.Code, request.RedirectURI)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to exchange OAuth code for JWT")
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
		h.log.Error().Err(err).Msg("Failed to bind refresh token request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(ctx.Request.Context(), request.RefreshToken)
	if err != nil {
		h.log.Debug().Err(err).Msg("Failed to refresh access token")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "failed to refresh access token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": accessToken})
}

// Login handles user login requests. It expects a JSON payload containing a username and password.
//
// On successful authentication, it returns an access token and a refresh token in the response body.
func (h *AuthAPIHandler) Login(ctx *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.log.Error().Err(err).Msg("Failed to bind login request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(ctx.Request.Context(), request.Username, request.Password)
	if err != nil {
		h.log.Debug().
			Err(err).
			Str("username", request.Username).
			Msg("Login attempt failed")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": accessToken, "refresh_token": refreshToken})
}
