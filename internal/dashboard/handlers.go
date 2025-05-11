package dashboard

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job"
	"github.com/benidevo/prospector/internal/job/models"
	"github.com/gin-gonic/gin"
)

// Handler manages dashboard related HTTP requests.
type Handler struct {
	cfg        *config.Settings
	jobService *job.JobService
}

// NewHandler creates and returns a new Handler.
func NewHandler(cfg *config.Settings, jobService *job.JobService) *Handler {
	return &Handler{
		cfg:        cfg,
		jobService: jobService,
	}
}

// GetDashboardPage renders the dashboard page template.
func (h *Handler) GetDashboardPage(c *gin.Context) {
	username, _ := c.Get("username")

	statusParam := c.Query("status")

	filter := models.JobFilter{
		Limit: 50,
	}

	if statusParam != "" && statusParam != "all" {
		jobStatus, err := models.JobStatusFromString(statusParam)
		if err == nil {
			filter.Status = &jobStatus
		}
	}

	if h.jobService == nil {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":        "Dashboard",
			"page":         "dashboard",
			"activeNav":    "jobs",
			"pageTitle":    "Job Matches",
			"currentYear":  time.Now().Year(),
			"username":     username,
			"jobs":         []*models.Job{},
			"statusFilter": statusParam,
		})
		return
	}

	jobs, err := h.jobService.GetJobs(c.Request.Context(), filter)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":        "Dashboard",
			"page":         "dashboard",
			"activeNav":    "jobs",
			"pageTitle":    "Job Matches",
			"currentYear":  time.Now().Year(),
			"username":     username,
			"jobs":         []*models.Job{},
			"statusFilter": statusParam,
		})
		return
	}

	appliedCount := 0
	for _, job := range jobs {
		if job.Status == models.APPLIED {
			appliedCount++
		}
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":        "Dashboard",
		"page":         "dashboard",
		"activeNav":    "jobs",
		"pageTitle":    "Job Matches",
		"currentYear":  time.Now().Year(),
		"username":     username,
		"jobs":         jobs,
		"totalJobs":    len(jobs),
		"applied":      appliedCount,
		"highMatch":    1, // Keeping this dummy data for now
		"statusFilter": statusParam,
	})
}
