package settings

import (
	"net/http"
	"strings"
	"time"

	"github.com/benidevo/ascentio/internal/settings/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SettingsHandler manages settings-related HTTP requests
type SettingsHandler struct {
	service              *SettingsService
	experienceHandler    *BaseSettingsHandler
	educationHandler     *BaseSettingsHandler
	certificationHandler *BaseSettingsHandler
}

// NewSettingsHandler creates a new SettingsHandler instance
func NewSettingsHandler(service *SettingsService) *SettingsHandler {
	experienceMetadata := EntityMetadata{
		Name:      "Experience",
		URLPrefix: "experience",
		CreateFunc: func() CRUDEntity {
			return &models.WorkExperience{}
		},
	}

	educationMetadata := EntityMetadata{
		Name:      "Education",
		URLPrefix: "education",
		CreateFunc: func() CRUDEntity {
			return &models.Education{}
		},
	}

	certificationMetadata := EntityMetadata{
		Name:      "Certification",
		URLPrefix: "certification",
		CreateFunc: func() CRUDEntity {
			return &models.Certification{}
		},
	}

	return &SettingsHandler{
		service:              service,
		experienceHandler:    NewBaseSettingsHandler(service, experienceMetadata),
		educationHandler:     NewBaseSettingsHandler(service, educationMetadata),
		certificationHandler: NewBaseSettingsHandler(service, certificationMetadata),
	}
}

// formatValidationError formats validation errors into user-friendly messages
func (h *SettingsHandler) formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			switch field {
			case "FirstName":
				if tag == "required" {
					return "First name is required"
				}
				if tag == "max" {
					return "First name must not exceed 100 characters"
				}
			case "LastName":
				if tag == "required" {
					return "Last name is required"
				}
				if tag == "max" {
					return "Last name must not exceed 100 characters"
				}
			case "PhoneNumber":
				if tag == "phone" {
					return "Phone number contains invalid characters or must be between 10-20 characters"
				}
			case "LinkedInProfile":
				if tag == "linkedin" {
					return "LinkedIn profile must be a valid LinkedIn URL"
				}
				if tag == "url" {
					return "LinkedIn profile must be a valid URL"
				}
			case "GitHubProfile":
				if tag == "github" {
					return "GitHub profile must be a valid GitHub URL"
				}
				if tag == "url" {
					return "GitHub profile must be a valid URL"
				}
			case "Website":
				if tag == "url" {
					return "Website must be a valid URL"
				}
			case "Skills":
				if tag == "max" {
					return "Skills must not exceed 50 items"
				}
				if tag == "dive" {
					return "Each skill must be between 1-100 characters"
				}
			default:
				switch tag {
				case "required":
					return field + " is required"
				case "min":
					return field + " is too short"
				case "max":
					return field + " is too long"
				case "url":
					return field + " must be a valid URL"
				default:
					return "Invalid " + strings.ToLower(field)
				}
			}
		}
	}
	return err.Error()
}

// GetProfileSettingsPage handles the request to display the profile settings page
func (h *SettingsHandler) GetProfileSettingsPage(c *gin.Context) {
	username, _ := c.Get("username")
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfileWithRelated(c.Request.Context(), userID.(int))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "general",
			"message": "Failed to load profile settings",
		})
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Profile",
		"page":           "settings-profile",
		"activeNav":      "profile",
		"activeSettings": "profile",
		"pageTitle":      "Profile",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"profile":        profile,
		"industries":     models.GetAllIndustries(),
	})
}

// HandleCreateProfile handles the creation or update of a user's profile settings
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
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
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
		// Check if it's a validation error
		if _, ok := err.(validator.ValidationErrors); ok {
			errorMessage := h.formatValidationError(err)
			c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
				"type":    "error",
				"context": "dashboard",
				"message": errorMessage,
			})
			return
		}
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Failed to update profile: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "dashboard",
		"message": "Personal information updated successfully",
	})
}

// HandleUpdateOnlineProfile handles the HTTP request to update a user's online profile links
func (h *SettingsHandler) HandleUpdateOnlineProfile(c *gin.Context) {
	userID := c.GetInt("userID")

	linkedInProfile := strings.TrimSpace(c.PostForm("linkedin_profile"))
	gitHubProfile := strings.TrimSpace(c.PostForm("github_profile"))
	website := strings.TrimSpace(c.PostForm("website"))

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Failed to load profile settings",
		})
		return
	}

	profile.LinkedInProfile = linkedInProfile
	profile.GitHubProfile = gitHubProfile
	profile.Website = website

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Failed to update online profiles",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "dashboard",
		"message": "Online profiles updated successfully",
	})
}

// HandleUpdateContext handles the HTTP request to update a user's personal context
func (h *SettingsHandler) HandleUpdateContext(c *gin.Context) {
	userID := c.GetInt("userID")

	context := strings.TrimSpace(c.PostForm("context"))

	if err := h.service.ValidateContext(context); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": err.Error(),
		})
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Failed to load profile settings",
		})
		return
	}

	profile.Context = context

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Failed to update personal context: ",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "dashboard",
		"message": "Personal context updated successfully",
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
		"title":          "Security",
		"page":           "settings-security",
		"activeNav":      "security",
		"activeSettings": "security",
		"pageTitle":      "Security",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"security":       security,
	})
}

// GetNotificationSettings handles the request to display the notification settings page
func (h *SettingsHandler) GetNotificationSettingsPage(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Notifications",
		"page":           "settings-notifications",
		"activeNav":      "notifications",
		"activeSettings": "notifications",
		"pageTitle":      "Notifications",
		"currentYear":    time.Now().Year(),
		"username":       username,
	})
}

// GetAddExperiencePage handles the HTTP request to render the page for adding a new experience.
func (h *SettingsHandler) GetAddExperiencePage(c *gin.Context) {
	h.experienceHandler.GetAddPage(c)
}

// GetEditExperiencePage handles the HTTP request to render the edit experience page.
func (h *SettingsHandler) GetEditExperiencePage(c *gin.Context) {
	h.experienceHandler.GetEditPage(c)
}

// HandleExperienceForm handles the HTTP request for creating a new experience entry.
func (h *SettingsHandler) HandleExperienceForm(c *gin.Context) {
	h.experienceHandler.HandleCreate(c)
}

// HandleUpdateExperienceForm handles the HTTP request for updating an experience form.
func (h *SettingsHandler) HandleUpdateExperienceForm(c *gin.Context) {
	h.experienceHandler.HandleUpdate(c)
}

// HandleDeleteWorkExperience handles the HTTP request to delete a work experience entry.
func (h *SettingsHandler) HandleDeleteWorkExperience(c *gin.Context) {
	h.experienceHandler.HandleDelete(c)
}

// GetAddEducationPage handles the HTTP request to display the page for adding a new education entry.
func (h *SettingsHandler) GetAddEducationPage(c *gin.Context) {
	h.educationHandler.GetAddPage(c)
}

// GetEditEducationPage handles the HTTP request to render the edit education page.
func (h *SettingsHandler) GetEditEducationPage(c *gin.Context) {
	h.educationHandler.GetEditPage(c)
}

// CreateEducationForm handles the HTTP request for creating a new education form.
func (h *SettingsHandler) CreateEducationForm(c *gin.Context) {
	h.educationHandler.HandleCreate(c)
}

// HandleUpdateEducationForm handles the HTTP request for updating an education form.
func (h *SettingsHandler) HandleUpdateEducationForm(c *gin.Context) {
	h.educationHandler.HandleUpdate(c)
}

// HandleDeleteEducation handles HTTP DELETE requests for deleting an education record.
func (h *SettingsHandler) HandleDeleteEducation(c *gin.Context) {
	h.educationHandler.HandleDelete(c)
}

// GetAddCertificationPage handles the HTTP request to render the page for adding a new certification.
func (h *SettingsHandler) GetAddCertificationPage(c *gin.Context) {
	h.certificationHandler.GetAddPage(c)
}

// GetEditCertificationPage handles the HTTP request to render the edit certification page.
func (h *SettingsHandler) GetEditCertificationPage(c *gin.Context) {
	h.certificationHandler.GetEditPage(c)
}

// CreateCertificationForm handles the HTTP request for creating a new certification form.
func (h *SettingsHandler) CreateCertificationForm(c *gin.Context) {
	h.certificationHandler.HandleCreate(c)
}

// HandleUpdateCertificationForm handles the HTTP request for updating a certification form.
func (h *SettingsHandler) HandleUpdateCertificationForm(c *gin.Context) {
	h.certificationHandler.HandleUpdate(c)
}

// HandleDeleteCertification handles HTTP DELETE requests for removing a certification.
func (h *SettingsHandler) HandleDeleteCertification(c *gin.Context) {
	h.certificationHandler.HandleDelete(c)
}
