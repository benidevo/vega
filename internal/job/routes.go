package job

import "github.com/gin-gonic/gin"

// RegisterRoutes registers job-related HTTP routes with the provided Gin router group.
func RegisterRoutes(router *gin.RouterGroup, handler *JobHandler) {
	router.GET("", handler.ListJobsPage)
	router.GET("/new", handler.GetNewJobForm)
	router.POST("/new", handler.CreateJob)

	jobRoutes := router.Group("")
	jobRoutes.Use(handler.ValidateJobID())
	{
		jobRoutes.GET("/:id/details", handler.GetJobDetails)
		jobRoutes.PUT("/:id/:field", handler.UpdateJobField)
		jobRoutes.DELETE("/:id", handler.DeleteJob)
		jobRoutes.POST("/:id/analyze", handler.AnalyzeJobMatch)
		jobRoutes.POST("/:id/cover-letter", handler.GenerateCoverLetter)
	}
}
