package prospector

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/auth"
	"github.com/benidevo/prospector/internal/home"
	"github.com/benidevo/prospector/internal/job"
	"github.com/benidevo/prospector/internal/settings"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes and middleware
func SetupRoutes(a *App) {
	authHandler := auth.SetupAuth(a.db, &a.config)
	homeHandler := home.Setup(&a.config)
	jobHandler := job.Setup(a.db, &a.config)
	settingsHandler := settings.Setup(&a.config)

	a.router.GET("/", homeHandler.GetHomePage)

	authGroup := a.router.Group("/auth")
	auth.RegisterPublicRoutes(authGroup, authHandler)
	authGroup.Use(authHandler.AuthMiddleware())
	auth.RegisterPrivateRoutes(authGroup, authHandler)

	jobGroup := a.router.Group("/jobs")
	jobGroup.Use(authHandler.AuthMiddleware())
	job.RegisterRoutes(jobGroup, jobHandler)

	settingsGroup := a.router.Group("/settings")
	settingsGroup.Use(authHandler.AuthMiddleware())
	settings.RegisterRoutes(settingsGroup, settingsHandler)

	a.router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "layouts/base.html", gin.H{
			"title":       "Page Not Found",
			"page":        "404",
			"currentYear": time.Now().Year(),
		})
	})
}
