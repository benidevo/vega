package services

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/benidevo/ascentio/internal/auth/models"
	"github.com/benidevo/ascentio/internal/auth/repository"
	"github.com/benidevo/ascentio/internal/common/logger"
	"github.com/benidevo/ascentio/internal/config"
	"github.com/rs/zerolog"
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
	log      zerolog.Logger
	repo     repository.UserRepository
}

// LogError logs an error from the Google authentication service
func (s *GoogleAuthService) LogError(err error) {
	s.log.Error().Err(err).Msg("Google authentication error")
}

// NewGoogleAuthService creates a new instance of GoogleAuthService using the provided configuration settings.
// It initializes the OAuth configuration and returns an error if credentials cannot be loaded.
func NewGoogleAuthService(cfg *config.Settings, repo repository.UserRepository) (*GoogleAuthService, error) {
	oauthCfg, err := getGoogleCredentials(cfg.GoogleClientConfigFile, cfg.GoogleClientRedirectURL, cfg.GoogleAuthUserInfoScope)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleCredentialsReadFailed, err)
	}

	return &GoogleAuthService{
		cfg:      cfg,
		oauthCfg: oauthCfg,
		log:      logger.GetLogger("google_auth"),
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
		s.log.Warn().Interface("user_info", userInfo).Msg("Google returned user info with empty email")
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
		Str("email", userInfo.Email).
		Bool("verified_email", userInfo.VerifiedEmail).
		Msg("Google user authenticated")

	user, err := s.getOrCreateUser(ctx, userInfo)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get or create Google user")
		return "", "", err
	}

	accessToken, err := GenerateAccessToken(user, s.cfg)
	if err != nil {
		s.log.Error().Err(err).Int("user_id", user.ID).Msg("Failed to generate access token for Google user")
		return "", "", models.ErrGoogleAuthTokenCreationFailed
	}

	refreshToken, err := GenerateRefreshToken(user, s.cfg)
	if err != nil {
		s.log.Error().Err(err).Int("user_id", user.ID).Msg("Failed to generate refresh token for Google user")
		return "", "", models.ErrGoogleAuthTokenCreationFailed
	}

	user.LastLogin = time.Now().UTC()
	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		sentinelErr := models.GetSentinelError(err)
		s.log.Warn().Err(err).Int("user_id", user.ID).Str("error_type", sentinelErr.Error()).Msg("Failed to update user last login time")
	}

	s.log.Info().Int("user_id", user.ID).Str("email", user.Username).Msg("Google user successfully logged in")
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
			s.log.Error().Err(err).Str("email", userInfo.Email).Str("error_type", sentinelErr.Error()).Msg("Error looking up Google user")
			return nil, models.ErrGoogleUserCreationFailed
		}

		// User doesn't exist, create a new one
		s.log.Info().Str("email", userInfo.Email).Msg("Creating new user from Google account")
		user, err = s.repo.CreateUser(ctx, userInfo.Email, "", models.STANDARD.String())
		if err != nil {
			sentinelErr := models.GetSentinelError(err)
			s.log.Error().Err(err).Str("email", userInfo.Email).Str("error_type", sentinelErr.Error()).Msg("Failed to create user from Google account")
			return nil, models.ErrGoogleUserCreationFailed
		}
		s.log.Info().Str("email", userInfo.Email).Int("user_id", user.ID).Msg("New user created from Google account")
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

// getGoogleCredentials reads a Google OAuth2 client secret JSON file from configPath,
// parses the credentials, and returns an oauth2.Config configured with the provided redirectURL.
//
// It returns an error if the file cannot be read or parsed.
func getGoogleCredentials(configPath, redirectURL, userInfoScope string) (*oauth2.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleCredentialsReadFailed, err)
	}
	var creds struct {
		Web struct {
			ClientID     string   `json:"client_id"`
			ClientSecret string   `json:"client_secret"`
			RedirectURIs []string `json:"redirect_uris"`
		} `json:"web"`
	}

	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrGoogleCredentialsInvalid, err)
	}

	if creds.Web.ClientID == "" || creds.Web.ClientSecret == "" {
		return nil, models.ErrGoogleCredentialsInvalid
	}

	oauthCfg := &oauth2.Config{
		ClientID:     creds.Web.ClientID,
		ClientSecret: creds.Web.ClientSecret,
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
