package job

import (
	"net/http"

	apimodels "github.com/benidevo/vega/internal/api/job/models"
	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/job/models"
	"github.com/gin-gonic/gin"
)

// JobAPIHandler handles job-related API requests
type JobAPIHandler struct {
	jobService *job.JobService
}

// NewJobAPIHandler creates a new job API handler
func NewJobAPIHandler(jobService *job.JobService) *JobAPIHandler {
	return &JobAPIHandler{
		jobService: jobService,
	}
}

// CreateJob handles the creation of a new job from API request
func (h *JobAPIHandler) CreateJob(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDValue.(int)

	var req apimodels.CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.jobService.LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	jobType := models.JobTypeFromString(req.JobType)

	jobOptions := []models.JobOption{
		models.WithLocation(req.Location),
		models.WithJobType(jobType),
		models.WithSourceURL(req.SourceURL),
		models.WithApplicationURL(req.ApplicationURL),
		models.WithNotes(req.Notes),
		models.WithStatus(models.INTERESTED),
	}

	createdJob, err := h.jobService.CreateJob(
		c.Request.Context(),
		userID,
		req.Title,
		req.Description,
		req.Company,
		jobOptions...,
	)

	if err != nil {
		switch err {
		case models.ErrJobTitleRequired,
			models.ErrJobDescriptionRequired,
			models.ErrCompanyRequired:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case models.ErrInvalidURLFormat:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid URL provided",
			})
		case models.ErrDuplicateJob:
			c.JSON(http.StatusConflict, gin.H{
				"error": "Job already exists with this source URL",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create job",
			})
		}
		return
	}

	c.JSON(http.StatusOK, apimodels.CreateJobResponse{
		Message: "Job created successfully",
		JobID:   createdJob.ID,
	})
}

// GetQuotaStatus returns the current quota status for the authenticated user
func (h *JobAPIHandler) GetQuotaStatus(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDValue.(int)

	quotaStatus, err := h.jobService.GetQuotaStatus(c.Request.Context(), userID)
	if err != nil {
		h.jobService.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get quota status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"used":  quotaStatus.Used,
		"limit": quotaStatus.Limit,
		"remaining": func() int {
			if quotaStatus.Limit < 0 {
				return -1 // Unlimited
			}
			return quotaStatus.Limit - quotaStatus.Used
		}(),
		"reset_date": quotaStatus.ResetDate.Format("2006-01-02"),
	})
}
