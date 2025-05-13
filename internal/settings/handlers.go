package settings

import (
	"net/http"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/gin-gonic/gin"
)

// SettingsHandler manages settings-related HTTP requests
type SettingsHandler struct {
	service *SettingsService
	cfg     *config.Settings
}

// NewSettingsHandler creates a new SettingsHandler instance
func NewSettingsHandler(service *SettingsService, cfg *config.Settings) *SettingsHandler {
	return &SettingsHandler{
		service: service,
		cfg:     cfg,
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
	userID := 1 // For now, hardcode userID

	profile, err := h.service.GetProfileSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to load profile settings",
		})
		return
	}

	workExperiences, _ := h.service.GetWorkExperiences(c.Request.Context(), profile.ID)
	education, _ := h.service.GetEducation(c.Request.Context(), profile.ID)
	certifications, _ := h.service.GetCertifications(c.Request.Context(), profile.ID)
	awards, _ := h.service.GetAwards(c.Request.Context(), profile.ID)

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Profile Settings",
		"page":           "settings-profile",
		"activeNav":      "settings",
		"activeSettings": "profile",
		"pageTitle":      "Profile Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"profile":        profile,
		"experiences":    workExperiences,
		"education":      education,
		"certifications": certifications,
		"awards":         awards,
	})
}

// GetSecuritySettings handles the request to display the security settings page
func (h *SettingsHandler) GetSecuritySettingsPage(c *gin.Context) {
	username, _ := c.Get("username")
	userID := 1 // For now, hardcode userID

	security, err := h.service.GetSecuritySettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to load security settings",
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
	userID := 1 // For now, hardcode userID

	notifications, err := h.service.GetNotificationSettings(c.Request.Context(), userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/alert-error.html", gin.H{
			"message": "Failed to load notification settings",
		})
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"title":          "Notification Settings",
		"page":           "settings-notifications",
		"activeNav":      "settings",
		"activeSettings": "notifications",
		"pageTitle":      "Notification Settings",
		"currentYear":    time.Now().Year(),
		"username":       username,
		"notifications":  notifications,
	})
}

// GetExperienceForm returns the form to add a new work experience
func (h *SettingsHandler) GetExperienceForm(c *gin.Context) {
	// For a new entry, I just return the empty form
	c.HTML(http.StatusOK, "settings/forms/experience_form.html", gin.H{})
}

// GetExperienceEditForm returns the form to edit an existing work experience
func (h *SettingsHandler) GetExperienceEditForm(c *gin.Context) {
	// For now I use dummy data
	c.HTML(http.StatusOK, "settings/forms/experience_form.html", gin.H{
		"experience": &struct {
			ID          int        `json:"id"`
			Title       string     `json:"title"`
			Company     string     `json:"company"`
			Location    string     `json:"location"`
			StartDate   time.Time  `json:"start_date"`
			EndDate     *time.Time `json:"end_date"`
			Description string     `json:"description"`
			Current     bool       `json:"current"`
		}{
			ID:          1,
			Title:       "Software Engineer",
			Company:     "Example Corp",
			Location:    "San Francisco, CA",
			StartDate:   time.Now().AddDate(-2, 0, 0),
			Description: "Worked on various projects",
			Current:     true,
		},
	})
}

// GetEducationForm returns the form to add a new education entry
func (h *SettingsHandler) GetEducationForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/education_form.html", gin.H{})
}

// GetEducationEditForm returns the form to edit an existing education entry
func (h *SettingsHandler) GetEducationEditForm(c *gin.Context) {
	// For a real implementation, I would fetch the education by ID
	// For now I'll use dummy data
	c.HTML(http.StatusOK, "settings/forms/education_form.html", gin.H{
		"education": &struct {
			ID           int        `json:"id"`
			Institution  string     `json:"institution"`
			Degree       string     `json:"degree"`
			FieldOfStudy string     `json:"field_of_study"`
			StartDate    time.Time  `json:"start_date"`
			EndDate      *time.Time `json:"end_date"`
			Description  string     `json:"description"`
		}{
			ID:           1,
			Institution:  "Stanford University",
			Degree:       "Bachelor of Science",
			FieldOfStudy: "Computer Science",
			StartDate:    time.Now().AddDate(-4, 0, 0),
			EndDate:      func() *time.Time { t := time.Now().AddDate(-1, 0, 0); return &t }(),
			Description:  "Graduated with honors",
		},
	})
}

// GetCertificationForm returns the form to add a new certification
func (h *SettingsHandler) GetCertificationForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/certification_form.html", gin.H{})
}

// GetCertificationEditForm returns the form to edit an existing certification
func (h *SettingsHandler) GetCertificationEditForm(c *gin.Context) {
	// For a real implementation, I would fetch the certification by ID
	// For now I'll use dummy data
	c.HTML(http.StatusOK, "settings/forms/certification_form.html", gin.H{
		"certification": &struct {
			ID            int        `json:"id"`
			Name          string     `json:"name"`
			IssuingOrg    string     `json:"issuing_org"`
			IssueDate     time.Time  `json:"issue_date"`
			ExpiryDate    *time.Time `json:"expiry_date"`
			CredentialID  string     `json:"credential_id"`
			CredentialURL string     `json:"credential_url"`
		}{
			ID:            1,
			Name:          "AWS Certified Developer",
			IssuingOrg:    "Amazon Web Services",
			IssueDate:     time.Now().AddDate(-1, 0, 0),
			ExpiryDate:    func() *time.Time { t := time.Now().AddDate(2, 0, 0); return &t }(),
			CredentialID:  "ABC123456",
			CredentialURL: "https://example.com/credentials/abc123456",
		},
	})
}

// GetAwardForm returns the form to add a new award
func (h *SettingsHandler) GetAwardForm(c *gin.Context) {
	c.HTML(http.StatusOK, "settings/forms/award_form.html", gin.H{})
}

// GetAwardEditForm returns the form to edit an existing award
func (h *SettingsHandler) GetAwardEditForm(c *gin.Context) {
	// For a real implementation, I would fetch the award by ID
	// For now I'll use dummy data
	c.HTML(http.StatusOK, "settings/forms/award_form.html", gin.H{
		"award": &struct {
			ID          int       `json:"id"`
			Title       string    `json:"title"`
			Issuer      string    `json:"issuer"`
			IssueDate   time.Time `json:"issue_date"`
			Description string    `json:"description"`
		}{
			ID:          1,
			Title:       "Employee of the Year",
			Issuer:      "Example Corp",
			IssueDate:   time.Now().AddDate(-1, 0, 0),
			Description: "Recognized for outstanding contributions to the team",
		},
	})
}

// CreateExperience handles form submission for adding work experience
func (h *SettingsHandler) CreateExperienceForm(c *gin.Context) {
	// In a real implementation, I would process the form data and create a new experience
	// For now, I'll just return a dummy updated list
	workExperiences := []*struct {
		ID          int        `json:"id"`
		Title       string     `json:"title"`
		Company     string     `json:"company"`
		Location    string     `json:"location"`
		StartDate   time.Time  `json:"start_date"`
		EndDate     *time.Time `json:"end_date"`
		Description string     `json:"description"`
		Current     bool       `json:"current"`
	}{
		{
			ID:          1,
			Title:       c.PostForm("title"),
			Company:     c.PostForm("company"),
			Location:    c.PostForm("location"),
			StartDate:   time.Now().AddDate(-2, 0, 0),
			Description: c.PostForm("description"),
			Current:     c.PostForm("current") == "on",
		},
	}

	c.Header("HX-Trigger", "closeModal")
	c.HTML(http.StatusOK, "partials/experience_list.html", gin.H{
		"experiences": workExperiences,
	})
}

// CreateEducation handles form submission for adding education
func (h *SettingsHandler) CreateEducationForm(c *gin.Context) {
	// In a real implementation, I would process the form data and create a new education entry
	// For now, I'll just return a dummy updated list
	educationList := []*struct {
		ID           int        `json:"id"`
		Institution  string     `json:"institution"`
		Degree       string     `json:"degree"`
		FieldOfStudy string     `json:"field_of_study"`
		StartDate    time.Time  `json:"start_date"`
		EndDate      *time.Time `json:"end_date"`
		Description  string     `json:"description"`
	}{
		{
			ID:           1,
			Institution:  c.PostForm("institution"),
			Degree:       c.PostForm("degree"),
			FieldOfStudy: c.PostForm("field_of_study"),
			StartDate:    time.Now().AddDate(-4, 0, 0),
			Description:  c.PostForm("description"),
		},
	}

	c.Header("HX-Trigger", "closeModal")
	c.HTML(http.StatusOK, "partials/education_list.html", gin.H{
		"education": educationList,
	})
}

// CreateCertification handles form submission for adding certification
func (h *SettingsHandler) CreateCertificationForm(c *gin.Context) {
	// In a real implementation, I would process the form data and create a new certification
	// For now, I'll just return a dummy updated list
	certifications := []*struct {
		ID            int        `json:"id"`
		Name          string     `json:"name"`
		IssuingOrg    string     `json:"issuing_org"`
		IssueDate     time.Time  `json:"issue_date"`
		ExpiryDate    *time.Time `json:"expiry_date"`
		CredentialID  string     `json:"credential_id"`
		CredentialURL string     `json:"credential_url"`
	}{
		{
			ID:            1,
			Name:          c.PostForm("name"),
			IssuingOrg:    c.PostForm("issuing_org"),
			IssueDate:     time.Now().AddDate(-1, 0, 0),
			CredentialID:  c.PostForm("credential_id"),
			CredentialURL: c.PostForm("credential_url"),
		},
	}

	c.Header("HX-Trigger", "closeModal")
	c.HTML(http.StatusOK, "partials/certification_list.html", gin.H{
		"certifications": certifications,
	})
}

// CreateAward handles form submission for adding an award
func (h *SettingsHandler) CreateAwardForm(c *gin.Context) {
	// In a real implementation, I would process the form data and create a new award
	// For now, I'll just return a dummy updated list
	awards := []*struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		Issuer      string    `json:"issuer"`
		IssueDate   time.Time `json:"issue_date"`
		Description string    `json:"description"`
	}{
		{
			ID:          1,
			Title:       c.PostForm("title"),
			Issuer:      c.PostForm("issuer"),
			IssueDate:   time.Now().AddDate(-1, 0, 0),
			Description: c.PostForm("description"),
		},
	}

	c.Header("HX-Trigger", "closeModal")
	c.HTML(http.StatusOK, "partials/award_list.html", gin.H{
		"awards": awards,
	})
}
