package settings

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benidevo/ascentio/internal/settings/models"
	"github.com/gin-gonic/gin"
)

// SettingsHandler manages settings-related HTTP requests
type SettingsHandler struct {
	service *SettingsService
}

// NewSettingsHandler creates a new SettingsHandler instance
func NewSettingsHandler(service *SettingsService) *SettingsHandler {
	return &SettingsHandler{
		service: service,
	}
}

// GetSettingsHome handles the request to display the settings home page
func (h *SettingsHandler) GetSettingsHomePage(c *gin.Context) {
	username, _ := c.Get("username")
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":       "Settings",
		"page":        "settings-home",
		"activeNav":   "settings",
		"pageTitle":   "Settings",
		"currentYear": time.Now().Year(),
		"username":    username,
	})
}

// GetProfileSettings handles the request to display the profile settings page
func (h *SettingsHandler) GetProfileSettingsPage(c *gin.Context) {
	username, _ := c.Get("username")
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID.(int))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Profile Settings",
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"profile":        profile,
	})
}

// HandleCreateProfile handles the creation or update of a user's profile settings.
func (h *SettingsHandler) HandleCreateProfile(c *gin.Context) {
	userID := c.GetInt("userID")

	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	title := strings.TrimSpace(c.PostForm("title"))
	industryStr := strings.TrimSpace(c.PostForm("industry"))
	location := strings.TrimSpace(c.PostForm("location"))
	phoneNumber := strings.TrimSpace(c.PostForm("phone_number"))
	careerSummary := strings.TrimSpace(c.PostForm("career_summary"))
	skillsStr := strings.TrimSpace(c.PostForm("skills"))

	var skills []string
	if skillsStr != "" {
		skillsParts := strings.Split(skillsStr, ",")
		for _, s := range skillsParts {
			skill := strings.TrimSpace(s)
			if skill != "" {
				skills = append(skills, skill)
			}
		}
	}

	industry := models.IndustryFromString(industryStr)

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error-dashboard.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	profile.FirstName = firstName
	profile.LastName = lastName
	profile.Title = title
	profile.Industry = industry
	profile.Location = location
	profile.PhoneNumber = phoneNumber
	profile.CareerSummary = careerSummary
	profile.Skills = skills

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error-dashboard.html", gin.H{
			"message": "Failed to update profile",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert-success-dashboard.html", gin.H{
		"message": "Personal information updated successfully",
	})
}

// HandleUpdateOnlineProfile handles the HTTP request to update a user's online profile links
// (LinkedIn, GitHub, and website).
func (h *SettingsHandler) HandleUpdateOnlineProfile(c *gin.Context) {
	userID := c.GetInt("userID")

	linkedInProfile := strings.TrimSpace(c.PostForm("linkedin_profile"))
	gitHubProfile := strings.TrimSpace(c.PostForm("github_profile"))
	website := strings.TrimSpace(c.PostForm("website"))

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error-dashboard.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	profile.LinkedInProfile = linkedInProfile
	profile.GitHubProfile = gitHubProfile
	profile.Website = website

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error-dashboard.html", gin.H{
			"message": "Failed to update online profiles",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert-success-dashboard.html", gin.H{
		"message": "Online profiles updated successfully",
	})
}

// GetSecuritySettings handles the request to display the security settings page
func (h *SettingsHandler) GetSecuritySettingsPage(c *gin.Context) {
	username, _ := c.Get("username")

	security, err := h.service.GetSecuritySettings(c.Request.Context(), username.(string))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title":       "Something Went Wrong",
			"page":        "500",
			"currentYear": time.Now().Year(),
		})
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Security Settings",
		"page":           "settings-security",
		"activeNav":      "settings",
		"activeSettings": "security",
		"pageTitle":      "Security Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"security":       security,
	})
}

// GetNotificationSettings handles the request to display the notification settings page
func (h *SettingsHandler) GetNotificationSettingsPage(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Notification Settings",
		"page":           "settings-notifications",
		"activeNav":      "settings",
		"activeSettings": "notifications",
		"pageTitle":      "Notification Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
	})
}

// GetExperienceForm returns the form to add a new work experience
func (h *SettingsHandler) GetExperienceForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/experience_form.html", gin.H{})
}

// HandleExperienceForm processes the submission of a work experience form,
// validates input fields, parses dates, retrieves the user's profile settings,
// creates a new work experience entry, and redirects to the profile settings page.
func (h *SettingsHandler) HandleExperienceForm(c *gin.Context) {
	userID := c.GetInt("userID")
	var err error

	jobTitle := strings.TrimSpace(c.PostForm("title"))
	company := strings.TrimSpace(c.PostForm("company"))
	location := strings.TrimSpace(c.PostForm("location"))
	startDate := strings.TrimSpace(c.PostForm("start_date"))
	endDate := strings.TrimSpace(c.PostForm("end_date"))
	description := strings.TrimSpace(c.PostForm("description"))
	current := strings.TrimSpace(c.PostForm("current")) == "on"

	var parsedStartDate time.Time
	if startDate != "" {
		parsedStartDate, err = time.Parse("2006-01", startDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid start date format. Please use YYYY-MM.",
			})
			return
		}
	}

	var parsedEndDate *time.Time
	if endDate != "" {
		t, err := time.Parse("2006-01", endDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid end date format. Please use YYYY-MM.",
			})
			return
		}
		parsedEndDate = &t
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	h.service.CreateWorkExperience(c.Request.Context(), &models.WorkExperience{
		Title:       jobTitle,
		ProfileID:   profile.ID,
		Company:     company,
		Location:    location,
		StartDate:   parsedStartDate,
		EndDate:     parsedEndDate,
		Description: description,
		Current:     current,
	})

	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// GetEducationForm returns the form to add a new education entry
func (h *SettingsHandler) GetEducationForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/education_form.html", gin.H{})
}

// GetCertificationForm returns the form to add a new certification
func (h *SettingsHandler) GetCertificationForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/certification_form.html", gin.H{})
}

// HandleEducationForm processes the submission of an education form,
// validates input fields, parses dates, retrieves the user's profile settings,
// creates a new education entry, and returns the updated education list.
func (h *SettingsHandler) CreateEducationForm(c *gin.Context) {
	userID := c.GetInt("userID")
	var err error

	institution := strings.TrimSpace(c.PostForm("institution"))
	degree := strings.TrimSpace(c.PostForm("degree"))
	fieldOfStudy := strings.TrimSpace(c.PostForm("field_of_study"))
	startDate := strings.TrimSpace(c.PostForm("start_date"))
	endDate := strings.TrimSpace(c.PostForm("end_date"))
	description := strings.TrimSpace(c.PostForm("description"))
	current := strings.TrimSpace(c.PostForm("current")) == "on"

	// Validate required fields
	if institution == "" || degree == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Institution and degree are required",
		})
		return
	}

	// Parse dates
	var parsedStartDate time.Time
	if startDate != "" {
		parsedStartDate, err = time.Parse("2006-01", startDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid start date format. Please use YYYY-MM.",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Start date is required",
		})
		return
	}

	var parsedEndDate *time.Time
	if !current && endDate != "" {
		t, err := time.Parse("2006-01", endDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid end date format. Please use YYYY-MM.",
			})
			return
		}
		parsedEndDate = &t
	}

	// Get user's profile
	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	// Create education entry
	education := &models.Education{
		ProfileID:    profile.ID,
		Institution:  institution,
		Degree:       degree,
		FieldOfStudy: fieldOfStudy,
		StartDate:    parsedStartDate,
		EndDate:      parsedEndDate,
		Description:  description,
	}

	err = h.service.CreateEducation(c.Request.Context(), education)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to create education entry: " + err.Error(),
		})
		return
	}

	c.Header("HX-Trigger", "closeModal")
	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// HandleUpdateEducationForm processes a request to update an existing education entry.
func (h *SettingsHandler) HandleUpdateEducationForm(c *gin.Context) {
	userID := c.GetInt("userID")
	educationIDStr := c.Param("id")
	var err error

	educationID, err := strconv.Atoi(educationIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid education ID format",
		})
		return
	}

	institution := strings.TrimSpace(c.PostForm("institution"))
	degree := strings.TrimSpace(c.PostForm("degree"))
	fieldOfStudy := strings.TrimSpace(c.PostForm("field_of_study"))
	startDate := strings.TrimSpace(c.PostForm("start_date"))
	endDate := strings.TrimSpace(c.PostForm("end_date"))
	description := strings.TrimSpace(c.PostForm("description"))
	current := strings.TrimSpace(c.PostForm("current")) == "on"

	// Validate required fields
	if institution == "" || degree == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Institution and degree are required",
		})
		return
	}

	// Parse dates
	var parsedStartDate time.Time
	if startDate != "" {
		parsedStartDate, err = time.Parse("2006-01", startDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid start date format. Please use YYYY-MM.",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Start date is required",
		})
		return
	}

	var parsedEndDate *time.Time
	if !current && endDate != "" {
		t, err := time.Parse("2006-01", endDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid end date format. Please use YYYY-MM.",
			})
			return
		}
		parsedEndDate = &t
	}

	// Get user's profile
	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	// Get the existing education entry to update
	education, err := h.service.GetEducationByID(c.Request.Context(), educationID, profile.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/alert-error.html", gin.H{
			"message": "Education entry not found or you don't have permission to edit it",
		})
		return
	}

	// Update education fields
	education.Institution = institution
	education.Degree = degree
	education.FieldOfStudy = fieldOfStudy
	education.StartDate = parsedStartDate
	education.EndDate = parsedEndDate
	education.Description = description

	if current {
		education.EndDate = nil
	}

	// Update the education entry
	err = h.service.UpdateEducation(c.Request.Context(), education)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to update education entry: " + err.Error(),
		})
		return
	}

	c.Header("HX-Trigger", "closeModal")
	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// HandleDeleteEducation handles the HTTP request to delete a user's education entry.
func (h *SettingsHandler) HandleDeleteEducation(c *gin.Context) {
	userID := c.GetInt("userID")
	educationIDStr := c.Param("id")

	educationID, err := strconv.Atoi(educationIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid education ID format",
		})
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	err = h.service.DeleteEducation(c.Request.Context(), educationID, profile.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to delete education entry: " + err.Error(),
		})
		return
	}

	c.String(http.StatusOK, "")
}

// CreateCertificationForm handles form submission for adding certification
func (h *SettingsHandler) CreateCertificationForm(c *gin.Context) {
	userID := c.GetInt("userID")
	var err error

	name := strings.TrimSpace(c.PostForm("name"))
	issuingOrg := strings.TrimSpace(c.PostForm("issuing_org"))
	issueDate := strings.TrimSpace(c.PostForm("issue_date"))
	expiryDate := strings.TrimSpace(c.PostForm("expiry_date"))
	credentialID := strings.TrimSpace(c.PostForm("credential_id"))
	credentialURL := strings.TrimSpace(c.PostForm("credential_url"))
	noExpiry := strings.TrimSpace(c.PostForm("no_expiry")) == "on"

	// Validate required fields
	if name == "" || issuingOrg == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Certification name and issuing organization are required",
		})
		return
	}

	// Parse dates
	var parsedIssueDate time.Time
	if issueDate != "" {
		parsedIssueDate, err = time.Parse("2006-01", issueDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid issue date format. Please use YYYY-MM.",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Issue date is required",
		})
		return
	}

	var parsedExpiryDate *time.Time
	if !noExpiry && expiryDate != "" {
		t, err := time.Parse("2006-01", expiryDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid expiry date format. Please use YYYY-MM.",
			})
			return
		}
		parsedExpiryDate = &t
	}

	// Get user's profile
	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	// Create certification
	certification := &models.Certification{
		ProfileID:     profile.ID,
		Name:          name,
		IssuingOrg:    issuingOrg,
		IssueDate:     parsedIssueDate,
		ExpiryDate:    parsedExpiryDate,
		CredentialID:  credentialID,
		CredentialURL: credentialURL,
	}

	err = h.service.CreateCertification(c.Request.Context(), certification)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to create certification: " + err.Error(),
		})
		return
	}

	c.Header("HX-Trigger", "closeModal")
	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// HandleUpdateCertificationForm processes a request to update an existing certification.
func (h *SettingsHandler) HandleUpdateCertificationForm(c *gin.Context) {
	userID := c.GetInt("userID")
	certificationIDStr := c.Param("id")
	var err error

	certificationID, err := strconv.Atoi(certificationIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid certification ID format",
		})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	issuingOrg := strings.TrimSpace(c.PostForm("issuing_org"))
	issueDate := strings.TrimSpace(c.PostForm("issue_date"))
	expiryDate := strings.TrimSpace(c.PostForm("expiry_date"))
	credentialID := strings.TrimSpace(c.PostForm("credential_id"))
	credentialURL := strings.TrimSpace(c.PostForm("credential_url"))
	noExpiry := strings.TrimSpace(c.PostForm("no_expiry")) == "on"

	// Validate required fields
	if name == "" || issuingOrg == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Certification name and issuing organization are required",
		})
		return
	}

	// Parse dates
	var parsedIssueDate time.Time
	if issueDate != "" {
		parsedIssueDate, err = time.Parse("2006-01", issueDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid issue date format. Please use YYYY-MM.",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Issue date is required",
		})
		return
	}

	var parsedExpiryDate *time.Time
	if !noExpiry && expiryDate != "" {
		t, err := time.Parse("2006-01", expiryDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid expiry date format. Please use YYYY-MM.",
			})
			return
		}
		parsedExpiryDate = &t
	}

	// Get user's profile
	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	// Get the existing certification to update
	certification, err := h.service.GetCertificationByID(c.Request.Context(), certificationID, profile.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/alert-error.html", gin.H{
			"message": "Certification not found or you don't have permission to edit it",
		})
		return
	}

	// Update certification fields
	certification.Name = name
	certification.IssuingOrg = issuingOrg
	certification.IssueDate = parsedIssueDate
	certification.ExpiryDate = parsedExpiryDate
	certification.CredentialID = credentialID
	certification.CredentialURL = credentialURL

	if noExpiry {
		certification.ExpiryDate = nil
	}

	// Update the certification
	err = h.service.UpdateCertification(c.Request.Context(), certification)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to update certification: " + err.Error(),
		})
		return
	}

	c.Header("HX-Trigger", "closeModal")
	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// HandleDeleteCertification handles the HTTP request to delete a user's certification.
func (h *SettingsHandler) HandleDeleteCertification(c *gin.Context) {
	userID := c.GetInt("userID")
	certificationIDStr := c.Param("id")

	certificationID, err := strconv.Atoi(certificationIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid certification ID format",
		})
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	err = h.service.DeleteCertification(c.Request.Context(), certificationID, profile.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to delete certification: " + err.Error(),
		})
		return
	}

	c.String(http.StatusOK, "")
}

// HandleUpdateExperienceForm processes a request to update an existing work experience.
// It parses form data, validates inputs, updates the specified work experience,
// and redirects back to the profile settings page.
func (h *SettingsHandler) HandleUpdateExperienceForm(c *gin.Context) {
	userID := c.GetInt("userID")
	experienceIDStr := c.Param("id")
	var err error

	experienceID, err := strconv.Atoi(experienceIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid experience ID format",
		})
		return
	}

	jobTitle := strings.TrimSpace(c.PostForm("title"))
	company := strings.TrimSpace(c.PostForm("company"))
	location := strings.TrimSpace(c.PostForm("location"))
	startDate := strings.TrimSpace(c.PostForm("start_date"))
	endDate := strings.TrimSpace(c.PostForm("end_date"))
	description := strings.TrimSpace(c.PostForm("description"))
	current := strings.TrimSpace(c.PostForm("current")) == "on"

	if jobTitle == "" || company == "" {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Job title and company name are required",
		})
		return
	}

	var parsedStartDate time.Time
	if startDate != "" {
		parsedStartDate, err = time.Parse("2006-01", startDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid start date format. Please use YYYY-MM.",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Start date is required",
		})
		return
	}

	var parsedEndDate *time.Time
	if !current && endDate != "" {
		t, err := time.Parse("2006-01", endDate)
		if err != nil {
			c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
				"message": "Invalid end date format. Please use YYYY-MM.",
			})
			return
		}
		parsedEndDate = &t
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	experience, err := h.service.GetWorkExperienceByID(c.Request.Context(), experienceID, profile.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/alert-error.html", gin.H{
			"message": "Work experience not found or you don't have permission to edit it",
		})
		return
	}
	experience.Title = jobTitle
	experience.Company = company
	experience.Location = location
	experience.StartDate = parsedStartDate
	experience.EndDate = parsedEndDate
	experience.Description = description
	experience.Current = current

	if current {
		experience.EndDate = nil
	}

	err = h.service.UpdateWorkExperience(c.Request.Context(), experience)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to update work experience: " + err.Error(),
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/settings/profile")
}

// HandleDeleteWorkExperience handles the HTTP request to delete a user's work experience entry.
// It validates the experience ID, retrieves the user's profile, and deletes the specified work experience.
// On success, it redirects to the profile settings page; on failure, it renders an error alert.
func (h *SettingsHandler) HandleDeleteWorkExperience(c *gin.Context) {
	userID := c.GetInt("userID")
	experienceIDStr := c.Param("id")

	experienceID, err := strconv.Atoi(experienceIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Invalid experience ID format",
		})
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	err = h.service.DeleteWorkExperience(c.Request.Context(), experienceID, profile.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to delete work experience: " + err.Error(),
		})
		return
	}

	c.String(http.StatusOK, "")
}
