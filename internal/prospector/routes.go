package prospector

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/auth"
	"github.com/benidevo/prospector/internal/home"
	"github.com/benidevo/prospector/internal/job"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes and middleware
func SetupRoutes(a *App) {
	authHandler := auth.SetupAuth(a.db, &a.config)
	homeHandler := home.Setup(&a.config)
	jobHandler := job.Setup(a.db, &a.config)

	a.router.GET("/", homeHandler.GetHomePage)

	authGroup := a.router.Group("/auth")
	{
		authGroup.GET("/login", authHandler.GetLoginPage)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.AuthMiddleware(), authHandler.Logout)
	}

	jobGroup := a.router.Group("/jobs", authHandler.AuthMiddleware())
	{
		jobGroup.GET("", jobHandler.ListJobsPage)
		jobGroup.GET("/new", jobHandler.GetNewJobForm)
		jobGroup.POST("/new", jobHandler.CreateJob)
		jobGroup.GET("/:id/details", jobHandler.GetJobDetails)
		jobGroup.PUT("/:id/:field", jobHandler.UpdateJobField)
		jobGroup.POST("/:id/:field", jobHandler.UpdateJobField) // Supports POST with X-HTTP-Method-Override header
		jobGroup.DELETE("/:id", jobHandler.DeleteJob)
	}
	a.router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "layouts/base.html", gin.H{
			"title":       "Page Not Found",
			"page":        "404",
			"currentYear": time.Now().Year(),
		})
	})
}
