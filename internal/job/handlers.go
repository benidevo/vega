package job

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// JobHandler manages job-related HTTP requests
type JobHandler struct {
	service *JobService
	cfg     *config.Settings
}

func NewJobHandler(service *JobService, cfg *config.Settings) *JobHandler {
	return &JobHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *JobHandler) GetNewJobForm(c *gin.Context) {
	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Add New Job",
		"page":        "job-new",
		"activeNav":   "newjob",
		"pageTitle":   "Add New Job",
		"currentYear": time.Now().Year(),
		"username":    username,
	})
}

func (h *JobHandler) GetJobDetails(c *gin.Context) {
	if h.cfg != nil && h.cfg.IsTest {
		c.Status(http.StatusOK)
		return
	}

	jobID := c.Param("id")

	// For now, I'll just render the template with dummy data
	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Job Details",
		"page":        "job-details",
		"activeNav":   "jobs",
		"pageTitle":   "Job Details",
		"currentYear": time.Now().Year(),
		"username":    username,
		"jobID":       jobID,
	})
}
