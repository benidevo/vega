package services

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/benidevo/vega/internal/auth/models"
	"github.com/benidevo/vega/internal/auth/repository"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/benidevo/vega/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GoogleAuthUserInfo represents the user information returned by Google's authentication API.
type GoogleAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

// GoogleAuthService handles authentication logic using Google's OAuth2 service.
type GoogleAuthService struct {
	cfg      *config.Settings
	oauthCfg *oauth2.Config
	log      *logger.PrivacyLogger
	repo     repository.UserRepository
}

// LogError logs an error from the Google authentication service
func (s *GoogleAuthService) LogError(err error) {
	s.log.Error().Err(err).Msg("Google authentication error")
}

// NewGoogleAuthService creates a new instance of GoogleAuthService using the provided configuration settings.
// It initializes the OAuth configuration and returns an error if credentials cannot be loaded.
func NewGoogleAuthService(cfg *config.Settings, repo repository.UserRepository) (*GoogleAuthService, error) {
	oauthCfg, err := getGoogleCredentials(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleClientRedirectURL, cfg.GoogleAuthUserInfoScope)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleCredentialsReadFailed, err)
	}

	return &GoogleAuthService{
		cfg:      cfg,
		oauthCfg: oauthCfg,
		log:      logger.GetPrivacyLogger("google_auth"),
		repo:     repo,
	}, nil
}

// GetAuthURL generates a Google OAuth2 authentication URL with a random state parameter.
// It requests offline access and forces the approval prompt.
func (s *GoogleAuthService) GetAuthURL() string {
	state := generateRandomState(32)
	return s.oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	)
}

// exchangeCode exchanges an authorization code for an OAuth2 token using the configured OAuth2 client.
func (s *GoogleAuthService) exchangeCode(ctx context.Context, code string, redirectURI string) (*oauth2.Token, error) {
	if redirectURI == "" {
		redirectURI = s.cfg.GoogleClientRedirectURL
	}

	cfg := &oauth2.Config{
		ClientID:     s.oauthCfg.ClientID,
		ClientSecret: s.oauthCfg.ClientSecret,
		Endpoint:     s.oauthCfg.Endpoint,
		Scopes:       s.oauthCfg.Scopes,
		RedirectURL:  redirectURI,
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		s.log.Error().Err(err).Str("code_length", fmt.Sprintf("%d", len(code))).Msg("Failed to exchange Google authcode")
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleCodeExchangeFailed, err)
	}
	return token, nil
}

// getUserInfo retrieves the authenticated user's information from Google using the provided OAuth2 token.
func (s *GoogleAuthService) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleAuthUserInfo, error) {
	client := s.oauthCfg.Client(ctx, token)

	resp, err := client.Get(s.cfg.GoogleAuthUserInfoURL)
	if err != nil {
		s.log.Error().Err(err).Str("url", s.cfg.GoogleAuthUserInfoURL).Msg("Failed to call Google UserInfo API")
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleUserInfoFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error().Int("status_code", resp.StatusCode).Str("status", resp.Status).Msg("Google UserInfo API returned non-OK status")
		return nil, fmt.Errorf("%w: %s", models.ErrGoogleUserInfoFailed, resp.Status)
	}

	var userInfo GoogleAuthUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		s.log.Error().Err(err).Msg("Failed to decode Google UserInfo response")
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleUserInfoFailed, err)
	}

	if userInfo.Email == "" {
		s.log.Warn().
			Str("event", "google_empty_email").
			Str("google_id", userInfo.ID).
			Msg("Google returned user info with empty email")
	}

	return &userInfo, nil
}

// Authenticate exchanges the provided Google OAuth code for access and refresh tokens.
// It retrieves user info, creates or fetches the user in the database, generates authentication tokens,
// updates the user's last login time, and returns the access and refresh tokens.
func (s *GoogleAuthService) Authenticate(ctx context.Context, code, redirect_uri string) (string, string, error) {
	token, err := s.exchangeCode(ctx, code, redirect_uri)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to exchange Google auth code")
		return "", "", err
	}

	userInfo, err := s.getUserInfo(ctx, token)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to retrieve Google user info")
		return "", "", err
	}

	s.log.Info().
		Str("event", "google_auth_success").
		Str("hashed_id", logger.HashIdentifier(userInfo.Email)).
		Bool("verified_email", userInfo.VerifiedEmail).
		Msg("Google user authenticated")

	user, err := s.getOrCreateUser(ctx, userInfo)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get or create Google user")
		return "", "", err
	}

	accessToken, err := GenerateAccessToken(user, s.cfg)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "google_access_token_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("Failed to generate access token for Google user")
		return "", "", models.ErrGoogleAuthTokenCreationFailed
	}

	refreshToken, err := GenerateRefreshToken(user, s.cfg)
	if err != nil {
		s.log.Error().Err(err).
			Str("event", "google_refresh_token_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("Failed to generate refresh token for Google user")
		return "", "", models.ErrGoogleAuthTokenCreationFailed
	}

	user.LastLogin = time.Now().UTC()
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Warn().Err(err).
			Str("event", "google_last_login_update_failed").
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Str("error_type", sentinelErr.Error()).
			Msg("Failed to update user last login time")
	}

	s.log.LogAuthEvent("google_login_success", user.ID, true)
	return accessToken, refreshToken, nil
}

// getOrCreateUser retrieves a user by their Google account email or creates a new user if not found.
// Returns the user or an error if the lookup or creation fails.
func (s *GoogleAuthService) getOrCreateUser(ctx context.Context, userInfo *GoogleAuthUserInfo) (*models.User, error) {
	var err error
	var user *models.User

	user, err = s.repo.FindByUsername(ctx, userInfo.Email)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)

		if sentinelErr != models.ErrUserNotFound {
			// This is an unexpected error
			s.log.Error().Err(err).
				Str("event", "google_user_lookup_error").
				Str("hashed_id", logger.HashIdentifier(userInfo.Email)).
				Str("error_type", sentinelErr.Error()).
				Msg("Error looking up Google user")
			return nil, models.ErrGoogleUserCreationFailed
		}

		// User doesn't exist, create a new one
		s.log.Info().
			Str("event", "google_user_create").
			Str("hashed_id", logger.HashIdentifier(userInfo.Email)).
			Msg("Creating new user from Google account")
		user, err = s.repo.CreateUser(ctx, userInfo.Email, "", models.STANDARD.String())
		if err != nil {
			sentinelErr := models.GetSentinelError(err)
			s.log.Error().Err(err).
				Str("event", "google_user_create_failed").
				Str("hashed_id", logger.HashIdentifier(userInfo.Email)).
				Str("error_type", sentinelErr.Error()).
				Msg("Failed to create user from Google account")
			return nil, models.ErrGoogleUserCreationFailed
		}
		s.log.Info().
			Str("event", "google_user_created").
			Str("hashed_id", logger.HashIdentifier(userInfo.Email)).
			Str("user_ref", fmt.Sprintf("user_%d", user.ID)).
			Msg("New user created from Google account")
	}

	return user, nil
}

// CreateDriveService creates a new Google Drive service client using the provided OAuth2 token and context.
func (s *GoogleAuthService) CreateDriveService(ctx context.Context, token *oauth2.Token) (*drive.Service, error) {
	client := s.oauthCfg.Client(ctx, token)

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create Google Drive service")
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleDriveServiceFailed, err)
	}

	s.log.Debug().Msg("Google Drive service created successfully")
	return driveService, nil
}

// getGoogleCredentials creates an oauth2.Config using environment variables
// for client ID and secret instead of reading from a JSON file.
//
// It returns an error if required credentials are missing.
func getGoogleCredentials(clientID, clientSecret, redirectURL, userInfoScope string) (*oauth2.Config, error) {
	if clientID == "" || clientSecret == "" {
		return nil, models.ErrGoogleCredentialsInvalid
	}

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			drive.DriveFileScope,
			userInfoScope,
		},
		Endpoint: google.Endpoint,
	}
	return oauthCfg, nil
}

// generateRandomState creates a random string of specified length for use as OAuth state parameter.
// This helps prevent CSRF attacks during the OAuth flow.
func generateRandomState(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	randomBytes := make([]byte, length)
	if _, err := cryptorand.Read(randomBytes); err != nil {
		// Fallback to a simple random string if crypto/rand fails
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return string(result)
	}

	for i, b := range randomBytes {
		result[i] = charset[int(b)%len(charset)]
	}

	return string(result)
}
