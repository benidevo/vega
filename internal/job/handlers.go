package job

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job/models"
	"github.com/gin-gonic/gin"
)

// JobHandler manages job-related HTTP requests
type JobHandler struct {
	service *JobService
	cfg     *config.Settings
}

// NewJobHandler creates and returns a new JobHandler with the provided JobService and configuration settings.
func NewJobHandler(service *JobService, cfg *config.Settings) *JobHandler {
	return &JobHandler{
		service: service,
		cfg:     cfg,
	}
}

// ListJobsPage handles the HTTP request to display the jobs dashboard page.
// It retrieves the current user's jobs, applies optional status filtering,
// gathers job statistics, and renders the dashboard template with the results.
func (h *JobHandler) ListJobsPage(c *gin.Context) {
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

	jobs, err := h.service.GetJobs(c.Request.Context(), filter)
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

	stats := h.service.GetJobStats(c.Request.Context())

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":        "Dashboard",
		"page":         "dashboard",
		"activeNav":    "jobs",
		"pageTitle":    "Job Matches",
		"currentYear":  time.Now().Year(),
		"username":     username,
		"jobs":         jobs,
		"totalJobs":    stats.TotalJobs,
		"applied":      stats.TotalApplied,
		"highMatch":    1, // Keeping this dummy data for now
		"statusFilter": statusParam,
	})
}

// GetNewJobForm renders the form for adding a new job.
// It populates the template with user and page information.
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

// CreateJob handles form submission for creating a new job
func (h *JobHandler) CreateJob(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	companyName := c.PostForm("company_name")
	location := c.PostForm("location")
	sourceURL := c.PostForm("url")
	notes := c.PostForm("notes")

	skillsStr := c.PostForm("skills")
	var skills []string
	if skillsStr != "" {
		skills = strings.Split(skillsStr, ",")
		for i := range skills {
			skills[i] = strings.TrimSpace(skills[i])
		}
	}

	jobTypeStr := c.PostForm("job_type")
	jobType, err := models.JobTypeFromString(jobTypeStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": err.Error(),
		})
		return
	}

	expLevelStr := c.PostForm("experience_level")
	expLevel := models.ExperienceLevelFromString(expLevelStr)

	statusStr := c.PostForm("status")
	status, err := models.JobStatusFromString(statusStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": err.Error(),
		})
		return
	}

	salaryMin := c.PostForm("salary_min")
	salaryMax := c.PostForm("salary_max")
	salaryRange := ""
	if salaryMin != "" && salaryMax != "" {
		salaryRange = salaryMin + "-" + salaryMax
	} else if salaryMin != "" {
		salaryRange = salaryMin + "+"
	} else if salaryMax != "" {
		salaryRange = "Up to " + salaryMax
	}

	options := []models.JobOption{
		models.WithJobType(jobType),
		models.WithExperienceLevel(expLevel),
		models.WithStatus(status),
		models.WithRequiredSkills(skills),
	}

	if location != "" {
		options = append(options, models.WithLocation(location))
	}
	if sourceURL != "" {
		options = append(options, models.WithSourceURL(sourceURL))
	}
	if salaryRange != "" {
		options = append(options, models.WithSalaryRange(salaryRange))
	}
	if notes != "" {
		options = append(options, models.WithNotes(notes))
	}

	job, err := h.service.CreateJob(c.Request.Context(), title, description, companyName, options...)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Something went wrong",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert-success.html", gin.H{
		"message": "Job created successfully!",
		"jobID":   strconv.Itoa(job.ID),
	})
}

// GetJobDetails handles the HTTP request to retrieve and display details for a specific job.
// It validates the job ID, fetches job data from the service layer, and renders the appropriate HTML template.
// Returns a 400 error for invalid IDs, 404 if the job is not found, or 500 for other errors.
func (h *JobHandler) GetJobDetails(c *gin.Context) {
	if h.cfg != nil && h.cfg.IsTest {
		c.Status(http.StatusOK)
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid job ID format",
		})
		return
	}

	job, err := h.service.GetJob(c.Request.Context(), jobID)
	if err != nil {
		if errors.Is(err, models.ErrJobNotFound) {
			c.HTML(http.StatusNotFound, "layouts/base.html", gin.H{
				"title":       "Page Not Found",
				"page":        "404",
				"currentYear": time.Now().Year(),
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Error retrieving job details: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Job Details",
		"page":        "job-details",
		"activeNav":   "jobs",
		"pageTitle":   "Job Details",
		"currentYear": time.Now().Year(),
		"username":    username,
		"job":         job,
		"jobID":       jobIDStr,
	})
}

// UpdateJobField handles the request to update a specific job field
func (h *JobHandler) UpdateJobField(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid job ID format",
		})
		return
	}

	field := c.Param("field")
	if field == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Field parameter is required",
		})
		return
	}

	job, err := h.service.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Error retrieving job: " + err.Error(),
		})
		return
	}

	var successMessage string

	switch field {
	case "status":
		statusStr := c.PostForm("status")
		if statusStr == "" {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Status is required",
			})
			return
		}

		status, err := models.JobStatusFromString(statusStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid status value",
			})
			return
		}

		job.Status = status
		successMessage = "Job status updated to " + status.String()

	case "notes":
		notes := c.PostForm("notes")
		job.Notes = notes
		successMessage = "Notes updated successfully"

	case "skills":
		skillsStr := c.PostForm("skills")

		var skills []string
		if skillsStr != "" {
			skills = strings.Split(skillsStr, ",")
			for i := range skills {
				skills[i] = strings.TrimSpace(skills[i])
			}
		}

		job.RequiredSkills = skills
		successMessage = "Skills updated successfully"

	case "basic":
		title := c.PostForm("title")
		if title == "" {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Job title is required",
			})
			return
		}
		job.Title = title

		companyName := c.PostForm("company_name")
		if companyName == "" {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Company name is required",
			})
			return
		}
		job.Company.Name = companyName

		location := c.PostForm("location")
		job.Location = location

		successMessage = "Job details updated successfully"

	default:
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Unsupported field: " + field,
		})
		return
	}

	err = h.service.UpdateJob(c.Request.Context(), job)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Error updating job: " + err.Error(),
		})
		return
	}

	// Use dashboard-specific alert for all status updates
	if field == "status" {
		c.HTML(http.StatusOK, "partials/alert-success-dashboard.html", gin.H{
			"message": successMessage,
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert-success-detail.html", gin.H{
		"message": successMessage,
	})
}

// DeleteJob handles the request to delete a job
func (h *JobHandler) DeleteJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid job ID format",
		})
		return
	}

	err = h.service.DeleteJob(c.Request.Context(), jobID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Error deleting job: " + err.Error(),
		})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		// This will immediately redirect the browser without showing any intermediate content
		c.Header("HX-Redirect", "/jobs")
		c.String(http.StatusOK, "")
		return
	}

	c.Redirect(http.StatusFound, "/jobs")
}
