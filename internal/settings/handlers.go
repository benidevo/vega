package settings

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/benidevo/vega/internal/ai"
	aimodels "github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/common/alerts"
	"github.com/benidevo/vega/internal/common/render"
	"github.com/benidevo/vega/internal/settings/models"
	"github.com/benidevo/vega/internal/settings/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SettingsHandler manages settings-related HTTP requests
type SettingsHandler struct {
	service              *SettingsService
	aiService            *ai.AIService
	experienceHandler    *BaseSettingsHandler
	educationHandler     *BaseSettingsHandler
	certificationHandler *BaseSettingsHandler
	renderer             *render.HTMLRenderer
}

// NewSettingsHandler creates a new SettingsHandler instance
func NewSettingsHandler(service *SettingsService, aiService *ai.AIService) *SettingsHandler {
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
		aiService:            aiService,
		experienceHandler:    NewBaseSettingsHandler(service, experienceMetadata),
		educationHandler:     NewBaseSettingsHandler(service, educationMetadata),
		certificationHandler: NewBaseSettingsHandler(service, certificationMetadata),
		renderer:             render.NewHTMLRenderer(service.cfg),
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
	userIDValue, exists := c.Get("userID")
	if !exists {
		h.renderer.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDValue.(int)

	profile, err := h.service.GetProfileWithRelated(c.Request.Context(), userID)

	if err != nil {
		h.renderer.HTML(c, http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title": "Something Went Wrong",
			"page":  "500",
		})
		return
	}

	if profile == nil {
		h.renderer.HTML(c, http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title": "Something Went Wrong",
			"page":  "500",
		})
		return
	}

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Profile",
		"page":           "settings-profile",
		"activeNav":      "profile",
		"activeSettings": "profile",
		"pageTitle":      "Profile",
		"profile":        profile,
		"industries":     models.GetAllIndustries(),
	})
}

// HandleCreateProfile handles the creation or update of a user's profile settings
func (h *SettingsHandler) HandleCreateProfile(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		h.renderer.HTML(c, http.StatusUnauthorized, "partials/alerts/alert.html", gin.H{
			"message": "Unauthorized",
			"type":    "error",
		})
		return
	}
	userID := userIDValue.(int)

	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	title := strings.TrimSpace(c.PostForm("title"))
	industryStr := strings.TrimSpace(c.PostForm("industry"))
	location := strings.TrimSpace(c.PostForm("location"))
	phoneNumber := strings.TrimSpace(c.PostForm("phone_number"))
	email := strings.TrimSpace(c.PostForm("email"))
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
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
		return
	}

	profile.FirstName = firstName
	profile.LastName = lastName
	profile.Title = title
	profile.Industry = industry
	profile.Location = location
	profile.PhoneNumber = phoneNumber
	profile.Email = email
	profile.CareerSummary = careerSummary
	profile.Skills = skills

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		// Check if it's a validation error
		if _, ok := err.(validator.ValidationErrors); ok {
			errorMessage := h.formatValidationError(err)
			alerts.RenderError(c, http.StatusBadRequest, errorMessage, alerts.ContextDashboard)
			return
		}
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to update profile: "+err.Error(), alerts.ContextDashboard)
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
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
		return
	}

	profile.LinkedInProfile = linkedInProfile
	profile.GitHubProfile = gitHubProfile
	profile.Website = website

	err = h.service.UpdateProfile(c.Request.Context(), profile)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to update online profiles", alerts.ContextDashboard)
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
		alerts.RenderError(c, http.StatusBadRequest, err.Error(), alerts.ContextDashboard)
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		alerts.RenderError(c, http.StatusInternalServerError, "Failed to load profile settings", alerts.ContextDashboard)
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

// HandleCVUpload handles the HTTP request to parse and save CV data
func (h *SettingsHandler) HandleCVUpload(c *gin.Context) {
	userID := c.GetInt("userID")

	var requestData struct {
		CVText string `json:"cv_text"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		h.service.log.Error().Err(err).Msg("Failed to parse CV upload request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	if requestData.CVText == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "CV text is required",
		})
		return
	}

	if h.aiService == nil {
		h.service.log.Error().Msg("AI service not available for CV parsing")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "CV parsing service is currently unavailable",
		})
		return
	}

	cvResult, err := h.aiService.CVParser.ParseCV(c.Request.Context(), requestData.CVText)
	if err != nil {
		h.service.log.Error().Err(err).Msg("Failed to parse CV with AI")

		if strings.Contains(err.Error(), "invalid document:") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to parse CV content",
			})
		}
		return
	}

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		profile = &models.Profile{UserID: userID}
	}

	if cvResult.PersonalInfo.FirstName != "" {
		profile.FirstName = cvResult.PersonalInfo.FirstName
	}
	if cvResult.PersonalInfo.LastName != "" {
		profile.LastName = cvResult.PersonalInfo.LastName
	}
	if cvResult.PersonalInfo.Title != "" {
		profile.Title = cvResult.PersonalInfo.Title
	}
	if cvResult.PersonalInfo.Phone != "" {
		profile.PhoneNumber = cvResult.PersonalInfo.Phone
	}
	if cvResult.PersonalInfo.Location != "" {
		profile.Location = cvResult.PersonalInfo.Location
	}
	if len(cvResult.Skills) > 0 {
		profile.Skills = cvResult.Skills
	}

	// Save profile
	if err := h.service.UpdateProfile(c.Request.Context(), profile); err != nil {
		h.service.log.Error().Err(err).Msg("Failed to save parsed CV data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update profile",
		})
		return
	}

	// Replace work experience only if CV contains experience data
	if len(cvResult.WorkExperience) > 0 {
		if err := h.service.DeleteAllWorkExperience(c.Request.Context(), profile.ID); err != nil {
			h.service.log.Error().Err(err).Msg("Failed to clear existing work experience")
		}

		for i, aiExp := range cvResult.WorkExperience {
			exp := h.convertAIWorkExperienceToModel(aiExp, profile.ID)
			if err := h.service.CreateEntity(c, &exp); err != nil {
				h.service.log.Error().Err(err).
					Int("experience_index", i).
					Str("company", exp.Company).
					Str("title", exp.Title).
					Msg("Failed to save work experience from AI-parsed CV")
			} else {
				h.service.log.Info().
					Str("company", exp.Company).
					Str("title", exp.Title).
					Msg("Successfully saved work experience from AI-parsed CV")
			}
		}
	}

	// Replace education only if CV contains education data
	if len(cvResult.Education) > 0 {
		if err := h.service.DeleteAllEducation(c.Request.Context(), profile.ID); err != nil {
			h.service.log.Error().Err(err).Msg("Failed to clear existing education")
		}

		for i, aiEdu := range cvResult.Education {
			edu := h.convertAIEducationToModel(aiEdu, profile.ID)
			if err := h.service.CreateEntity(c, &edu); err != nil {
				h.service.log.Error().Err(err).
					Int("education_index", i).
					Str("institution", edu.Institution).
					Str("degree", edu.Degree).
					Msg("Failed to save education from AI-parsed CV")
			} else {
				h.service.log.Info().
					Str("institution", edu.Institution).
					Str("degree", edu.Degree).
					Msg("Successfully saved education from AI-parsed CV")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "CV parsed and profile updated successfully",
	})
}

// convertAIWorkExperienceToModel converts AI-parsed work experience to database model
func (h *SettingsHandler) convertAIWorkExperienceToModel(aiExp aimodels.WorkExperience, profileID int) models.WorkExperience {
	startDate := h.parseAIDate(aiExp.StartDate)
	var endDate *time.Time
	current := false

	if aiExp.EndDate != "" && strings.ToLower(aiExp.EndDate) != "present" {
		endDateTime := h.parseAIDate(aiExp.EndDate)
		endDate = &endDateTime
	} else if strings.ToLower(aiExp.EndDate) == "present" {
		current = true
	}

	return models.WorkExperience{
		ProfileID:   profileID,
		Company:     aiExp.Company,
		Title:       aiExp.Title,
		Location:    aiExp.Location,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: aiExp.Description,
		Current:     current,
	}
}

// convertAIEducationToModel converts AI-parsed education to database model
func (h *SettingsHandler) convertAIEducationToModel(aiEdu aimodels.Education, profileID int) models.Education {
	startDate := h.parseAIDate(aiEdu.StartDate)
	var endDate *time.Time

	if aiEdu.EndDate != "" {
		endDateTime := h.parseAIDate(aiEdu.EndDate)
		endDate = &endDateTime
	}

	return models.Education{
		ProfileID:    profileID,
		Institution:  aiEdu.Institution,
		Degree:       aiEdu.Degree,
		FieldOfStudy: aiEdu.FieldOfStudy,
		StartDate:    startDate,
		EndDate:      endDate,
	}
}

// parseAIDate parses AI-formatted dates (YYYY-MM or YYYY) to time.Time
func (h *SettingsHandler) parseAIDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now().AddDate(-4, 0, 0) // Default to 4 years ago
	}

	// Try YYYY-MM format first
	if t, err := time.Parse("2006-01", dateStr); err == nil {
		return t
	}

	// Try YYYY format
	if t, err := time.Parse("2006", dateStr); err == nil {
		return t
	}

	// Default fallback
	return time.Now().AddDate(-4, 0, 0)
}

// GetSecuritySettings handles the request to display the security settings page
func (h *SettingsHandler) GetSecuritySettingsPage(c *gin.Context) {
	// Return 404 if security page is disabled
	if !h.service.cfg.SecurityPageEnabled {
		h.renderer.HTML(c, http.StatusNotFound, "layouts/base.html", gin.H{
			"title": "Page Not Found",
			"page":  "404",
		})
		return
	}

	username, _ := c.Get("username")

	security, err := h.service.GetSecuritySettings(c.Request.Context(), username.(string))
	if err != nil {
		h.renderer.HTML(c, http.StatusInternalServerError, "layouts/base.html", gin.H{
			"title": "Something Went Wrong",
			"page":  "500",
		})
		return
	}

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Security",
		"page":           "settings-security",
		"activeNav":      "security",
		"activeSettings": "security",
		"pageTitle":      "Security",
		"security":       security,
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

// Job Search Preference Handlers

// GetJobSearchPreferencesPage displays the job search preferences page
func (h *SettingsHandler) GetJobSearchPreferencesPage(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		h.renderer.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDValue.(int)

	preferences, err := h.service.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		h.renderer.Error(c, http.StatusInternalServerError, "Failed to retrieve preferences")
		return
	}

	// Pass MaxAge constants to template - using slice of structs to maintain order
	type maxAgeOption struct {
		Label    string
		Value    int
		Selected bool
	}

	maxAgeOptions := []maxAgeOption{
		{"1 hour", models.MaxAgeOneHour, false},
		{"6 hours", models.MaxAgeSixHours, false},
		{"12 hours", models.MaxAgeTwelveHours, false},
		{"1 day", models.MaxAgeOneDay, true},
		{"3 days", models.MaxAgeThreeDays, false},
		{"1 week", models.MaxAgeOneWeek, false},
		{"2 weeks", models.MaxAgeTwoWeeks, false},
		{"30 days", models.MaxAgeThirtyDays, false},
	}

	data := gin.H{
		"title":          "Search Preferences",
		"activeNav":      "search-preferences",
		"page":           "settings-search-preferences",
		"activeSettings": "search-preferences",
		"pageTitle":      "Search Preferences",
		"preferences":    preferences,
		"maxAgeOptions":  maxAgeOptions,
		"maxPreferences": services.MaxPreferencesPerUser,
	}

	h.renderer.HTML(c, http.StatusOK, "layouts/base.html", data)
}

// CreateJobSearchPreference handles creating a new job search preference
func (h *SettingsHandler) CreateJobSearchPreference(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.HTML(http.StatusUnauthorized, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Unauthorized",
		})
		return
	}
	userID := userIDValue.(int)

	// Parse form data manually to handle type conversions
	jobTitle := strings.TrimSpace(c.PostForm("job_title"))
	location := strings.TrimSpace(c.PostForm("location"))
	skillsStr := strings.TrimSpace(c.PostForm("skills"))
	maxAgeStr := c.PostForm("max_age")
	isActiveStr := c.PostForm("is_active")

	// Validate required fields
	if jobTitle == "" || location == "" || maxAgeStr == "" {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Please fill in all required fields",
		})
		return
	}

	// Parse skills from comma-separated string
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

	// Convert max_age to int
	maxAge := 0
	if _, err := fmt.Sscanf(maxAgeStr, "%d", &maxAge); err != nil {
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": "Invalid maximum job age",
		})
		return
	}

	// Handle checkbox - if present it's checked, if not it's unchecked
	isActive := isActiveStr == "on"

	input := services.CreatePreferenceInput{
		JobTitle: jobTitle,
		Location: location,
		Skills:   skills,
		MaxAge:   maxAge,
		IsActive: isActive,
	}

	_, err := h.service.CreatePreference(c.Request.Context(), userID, input)
	if err != nil {
		var message string
		if err == services.ErrMaxPreferencesReached {
			message = "Maximum number of preferences reached (4)"
		} else if err == services.ErrInvalidPreferenceData {
			message = "Invalid preference data. Please check your inputs."
		} else {
			message = "Failed to create preference"
		}
		c.HTML(http.StatusBadRequest, "partials/alert.html", gin.H{
			"type":    "error",
			"context": "dashboard",
			"message": message,
		})
		return
	}

	// Redirect with success
	c.Header("HX-Redirect", "/settings/search-preferences")
	c.HTML(http.StatusOK, "partials/alert.html", gin.H{
		"type":    "success",
		"context": "dashboard",
		"message": "Job search preference created successfully",
	})
}

// UpdateJobSearchPreference handles updating an existing job search preference
func (h *SettingsHandler) UpdateJobSearchPreference(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)
	preferenceID := c.Param("id")

	var input services.UpdatePreferenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	err := h.service.UpdatePreference(c.Request.Context(), userID, preferenceID, input)
	if err != nil {
		if err == services.ErrPreferenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		} else if err == services.ErrInvalidPreferenceData {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference data"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preference"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference updated successfully"})
}

// DeleteJobSearchPreference handles deleting a job search preference
func (h *SettingsHandler) DeleteJobSearchPreference(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)
	preferenceID := c.Param("id")

	err := h.service.DeletePreference(c.Request.Context(), userID, preferenceID)
	if err != nil {
		if err == services.ErrPreferenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete preference"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference deleted successfully"})
}

// ToggleJobSearchPreference handles toggling the active status of a preference
func (h *SettingsHandler) ToggleJobSearchPreference(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDValue.(int)
	preferenceID := c.Param("id")

	err := h.service.TogglePreferenceStatus(c.Request.Context(), userID, preferenceID)
	if err != nil {
		if err == services.ErrPreferenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle preference status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference status toggled successfully"})
}
