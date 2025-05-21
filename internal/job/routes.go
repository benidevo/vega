package job

import "github.com/gin-gonic/gin"

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
