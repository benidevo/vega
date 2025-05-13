package job

import (
	"database/sql"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job/interfaces"
	"github.com/benidevo/prospector/internal/job/repository"
	"github.com/gin-gonic/gin"
)

// Setup initializes the job package dependencies and returns a JobHandler.
func Setup(db *sql.DB, cfg *config.Settings) *JobHandler {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	jobRepo := repository.NewSQLiteJobRepository(db, companyRepo)
	service := NewJobService(jobRepo, cfg)

	return NewJobHandler(service, cfg)
}

// SetupJobRepository initializes and returns a job repository.
func SetupJobRepository(db *sql.DB) interfaces.JobRepository {
	companyRepo := repository.NewSQLiteCompanyRepository(db)
	return repository.NewSQLiteJobRepository(db, companyRepo)
}

// RegisterRoutes registers job-related HTTP routes with the provided Gin router group.
func RegisterRoutes(router *gin.RouterGroup, handler *JobHandler) {
	router.GET("", handler.ListJobsPage)
	router.GET("/new", handler.GetNewJobForm)
	router.POST("/new", handler.CreateJob)
	router.GET("/:id/details", handler.GetJobDetails)
	router.PUT("/:id/:field", handler.UpdateJobField)
	router.POST("/:id/:field", handler.UpdateJobField) // Supports POST with X-HTTP-Method-Override header
	router.DELETE("/:id", handler.DeleteJob)
}
