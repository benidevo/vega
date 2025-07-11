package vega

import (
	"net/http"
	"time"

	"github.com/benidevo/vega/internal/ai"
	authapi "github.com/benidevo/vega/internal/api/auth"
	jobapi "github.com/benidevo/vega/internal/api/job"
	"github.com/benidevo/vega/internal/auth"
	"github.com/benidevo/vega/internal/home"
	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/settings"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SetupRoutes configures all application routes and middleware
func SetupRoutes(a *App) {
	a.router.Use(globalErrorHandler)

	a.router.Static("/static", "./static")

	// health check
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	aiService, err := ai.Setup(&a.config)
	if err != nil {
		log.Warn().Err(err).Msg("AI service initialization failed, AI features will be disabled")
		aiService = nil
	}

	authHandler := auth.SetupAuth(a.db, &a.config)
	homeHandler := home.Setup(a.db, &a.config)
	jobHandler := job.Setup(a.db, &a.config)
	settingsHandler := settings.Setup(&a.config, a.db, aiService)
	authAPIHandler := authapi.Setup(a.db, &a.config)
	jobAPIHandler := jobapi.Setup(a.db, &a.config)

	authGroup := a.router.Group("/auth")

	// Register password-based auth routes only if not in OAuth-only mode
	if !a.config.OAuthOnlyMode {
		auth.RegisterPublicRoutes(authGroup, authHandler)
	}

	// Setup and register Google Auth routes only if enabled
	if a.config.GoogleOAuthEnabled {
		googleAuthHandler, err := auth.SetupGoogleAuth(&a.config, a.db)
		if err != nil {
			log.Error().Err(err).Msg("Failed to setup Google Auth")
		} else {
			auth.RegisterGoogleAuthRoutes(authGroup, googleAuthHandler)
		}
	}

	authGroup.Use(authHandler.AuthMiddleware())
	auth.RegisterPrivateRoutes(authGroup, authHandler)

	// Homepage route with optional auth (accessible to all, populates context when authenticated)
	a.router.GET("/", authHandler.OptionalAuthMiddleware(), homeHandler.GetHomePage)

	jobGroup := a.router.Group("/jobs")
	jobGroup.Use(authHandler.AuthMiddleware())
	job.RegisterRoutes(jobGroup, jobHandler)

	settingsGroup := a.router.Group("/settings")
	settingsGroup.Use(authHandler.AuthMiddleware())
	settings.RegisterRoutes(settingsGroup, settingsHandler)

	authAPIGroup := a.router.Group("/api/auth")
	authapi.RegisterRoutes(authAPIGroup, authAPIHandler)

	jobAPIGroup := a.router.Group("/api/jobs")
	jobAPIGroup.Use(authHandler.APIAuthMiddleware())
	jobapi.RegisterRoutes(jobAPIGroup, jobAPIHandler)

	a.router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "layouts/base.html", gin.H{
			"title":       "Page Not Found",
			"page":        "404",
			"currentYear": time.Now().Year(),
		})
	})
}

// globalErrorHandler is a Gin middleware that recovers from panics and handles internal server errors
// by rendering a generic 500 error page.
//
// It ensures that any unhandled errors or panics result in a
// consistent error response to the client.
func globalErrorHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Err(err.(error)).Msg("Recovered from panic")

			c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
				"title":       "Something Went Wrong",
				"page":        "500",
				"currentYear": time.Now().Year(),
			})
			c.Abort()
		}
	}()

	c.Next()

	// Only handle errors if no response has been written yet
	if !c.Writer.Written() && (len(c.Errors) > 0 || c.Writer.Status() == http.StatusInternalServerError) {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":       "Something Went Wrong",
			"page":        "500",
			"currentYear": time.Now().Year(),
		})
		c.Abort()
	}
}
