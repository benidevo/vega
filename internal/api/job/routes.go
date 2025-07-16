package job

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes configures the API routes for job endpoints
func RegisterRoutes(router *gin.RouterGroup, handler *JobAPIHandler) {
	jobRoutes := router.Group("")
	{
		jobRoutes.POST("", handler.CreateJob)
		jobRoutes.GET("/quota", handler.GetQuotaStatus)
	}
}
