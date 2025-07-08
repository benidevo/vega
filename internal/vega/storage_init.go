package vega

import (
	"context"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/storage"
	"github.com/benidevo/vega/internal/storage/badger"
	"github.com/benidevo/vega/internal/storage/drive"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// InitializeStorageProviders sets up the storage providers based on configuration
func InitializeStorageProviders(app *App) error {
	if !app.config.MultiTenancyEnabled {
		return nil
	}

	cacheDir := "/app/data/cache"
	if app.config.IsDevelopment {
		cacheDir = "./data/cache"
	}

	// Always create Badger provider as cache layer
	cacheProvider := badger.NewProvider(cacheDir)

	if app.config.GoogleDriveStorage && app.config.GoogleOAuthEnabled {
		// Create OAuth2 config for Google Drive
		oauth2Config := &oauth2.Config{
			ClientID:     app.config.GoogleClientID,
			ClientSecret: app.config.GoogleClientSecret,
			RedirectURL:  app.config.GoogleClientRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/drive.appdata", // App-specific storage
			},
			Endpoint: google.Endpoint,
		}

		// Create Drive provider with cache
		driveProvider := drive.NewProvider(cacheProvider, oauth2Config)
		storageProvider := &driveStorageProvider{
			driveProvider: driveProvider,
			cacheProvider: cacheProvider,
			oauth2Config:  oauth2Config,
		}

		app.storageFactory.SetProvider(storageProvider)

		log.Info().
			Str("cache_dir", cacheDir).
			Bool("google_drive", true).
			Msg("Initialized storage with Google Drive backend")
	} else {
		app.storageFactory.SetProvider(cacheProvider)
		log.Info().Str("cache_dir", cacheDir).Msg("Initialized Badger storage provider")
	}
	return nil
}

// driveStorageProvider wraps the drive provider to handle OAuth tokens
type driveStorageProvider struct {
	driveProvider *drive.Provider
	cacheProvider storage.StorageProvider
	oauth2Config  *oauth2.Config
}

// GetStorage returns storage instance with Google Drive support
func (p *driveStorageProvider) GetStorage(ctx context.Context, userID string) (storage.UserStorage, error) {
	// Try to get OAuth token from context
	c, ok := ctx.Value("gin_context").(*gin.Context)
	if !ok {
		// No gin context, fall back to cache only
		return p.cacheProvider.GetStorage(ctx, userID)
	}

	// Get OAuth token from session or auth context
	tokenInterface, exists := c.Get("oauth_token")
	if !exists {
		// No token, fall back to cache only
		return p.cacheProvider.GetStorage(ctx, userID)
	}

	token, ok := tokenInterface.(*oauth2.Token)
	if !ok || token == nil {
		// Invalid token, fall back to cache only
		return p.cacheProvider.GetStorage(ctx, userID)
	}

	// Get Drive storage with token
	return p.driveProvider.GetStorage(ctx, userID, token)
}

// CloseAll closes all storage instances
func (p *driveStorageProvider) CloseAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return p.driveProvider.Close(ctx)
}

// UpdateTenantMiddleware updates the tenant middleware to include OAuth token in context
func UpdateTenantMiddleware(factory *storage.Factory, cfg *config.Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tenant isolation if not in cloud mode
		if !cfg.IsCloudMode || !cfg.MultiTenancyEnabled {
			c.Next()
			return
		}

		// Get user from context set by auth middleware
		username, exists := c.Get("username")
		if !exists {
			c.Next()
			return
		}

		userEmail, ok := username.(string)
		if !ok || userEmail == "" {
			c.Next()
			return
		}

		// Create context with gin context for OAuth token access
		ctx := context.WithValue(c.Request.Context(), "gin_context", c)

		userStorage, err := factory.GetUserStorage(ctx, userEmail)
		if err != nil {
			c.HTML(500, "layouts/base.html", gin.H{
				"title":       "Something Went Wrong",
				"page":        "500",
				"currentYear": 2025,
			})
			c.Abort()
			return
		}

		c.Set("user_storage", userStorage)
		c.Next()
	}
}
