package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// parseDate parses a date string in YYYY-MM format
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	return time.Parse("2006-01", dateStr)
}

// parseOptionalDate parses an optional date string
func parseOptionalDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01", dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// getBoolFromForm gets a boolean value from form checkbox
func getBoolFromForm(c *gin.Context, key string) bool {
	return strings.TrimSpace(c.PostForm(key)) == "on"
}

// BindFromForm implements FormBindable for WorkExperience
func (w *WorkExperience) BindFromForm(c *gin.Context) error {
	w.Title = strings.TrimSpace(c.PostForm("title"))
	w.Company = strings.TrimSpace(c.PostForm("company"))
	w.Location = strings.TrimSpace(c.PostForm("location"))
	w.Description = strings.TrimSpace(c.PostForm("description"))
	w.Current = getBoolFromForm(c, "current")

	// Validate required fields
	if w.Title == "" || w.Company == "" {
		return fmt.Errorf("job title and company name are required")
	}

	// Parse start date
	startDateStr := strings.TrimSpace(c.PostForm("start_date"))
	startDate, err := parseDate(startDateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format. Please use YYYY-MM")
	}
	w.StartDate = startDate

	// Parse end date
	endDateStr := strings.TrimSpace(c.PostForm("end_date"))
	if !w.Current && endDateStr != "" {
		endDate, err := parseOptionalDate(endDateStr)
		if err != nil {
			return fmt.Errorf("invalid end date format. Please use YYYY-MM")
		}
		w.EndDate = endDate
	} else if w.Current {
		w.EndDate = nil
	}

	return nil
}

// BindFromForm implements FormBindable for Education
func (e *Education) BindFromForm(c *gin.Context) error {
	e.Institution = strings.TrimSpace(c.PostForm("institution"))
	e.Degree = strings.TrimSpace(c.PostForm("degree"))
	e.FieldOfStudy = strings.TrimSpace(c.PostForm("field_of_study"))
	e.Description = strings.TrimSpace(c.PostForm("description"))
	current := getBoolFromForm(c, "current")

	// Validate required fields
	if e.Institution == "" || e.Degree == "" {
		return fmt.Errorf("institution and degree are required")
	}

	// Parse start date
	startDateStr := strings.TrimSpace(c.PostForm("start_date"))
	startDate, err := parseDate(startDateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format. Please use YYYY-MM")
	}
	e.StartDate = startDate

	// Parse end date
	endDateStr := strings.TrimSpace(c.PostForm("end_date"))
	if !current && endDateStr != "" {
		endDate, err := parseOptionalDate(endDateStr)
		if err != nil {
			return fmt.Errorf("invalid end date format. Please use YYYY-MM")
		}
		e.EndDate = endDate
	} else if current {
		e.EndDate = nil
	}

	return nil
}

// BindFromForm implements FormBindable for Certification
func (cert *Certification) BindFromForm(c *gin.Context) error {
	cert.Name = strings.TrimSpace(c.PostForm("name"))
	cert.IssuingOrg = strings.TrimSpace(c.PostForm("issuing_org"))
	cert.CredentialID = strings.TrimSpace(c.PostForm("credential_id"))
	cert.CredentialURL = strings.TrimSpace(c.PostForm("credential_url"))
	noExpiry := getBoolFromForm(c, "no_expiry")

	// Validate required fields
	if cert.Name == "" || cert.IssuingOrg == "" {
		return fmt.Errorf("certification name and issuing organization are required")
	}

	// Parse issue date
	issueDateStr := strings.TrimSpace(c.PostForm("issue_date"))
	issueDate, err := parseDate(issueDateStr)
	if err != nil {
		return fmt.Errorf("invalid issue date format. Please use YYYY-MM")
	}
	cert.IssueDate = issueDate

	// Parse expiry date
	expiryDateStr := strings.TrimSpace(c.PostForm("expiry_date"))
	if !noExpiry && expiryDateStr != "" {
		expiryDate, err := parseOptionalDate(expiryDateStr)
		if err != nil {
			return fmt.Errorf("invalid expiry date format. Please use YYYY-MM")
		}
		cert.ExpiryDate = expiryDate
	} else if noExpiry {
		cert.ExpiryDate = nil
	}

	return nil
}

// GetID implements CRUDEntity for WorkExperience
func (w *WorkExperience) GetID() int {
	return w.ID
}

// GetProfileID implements CRUDEntity for WorkExperience
func (w *WorkExperience) GetProfileID() int {
	return w.ProfileID
}

// GetID implements CRUDEntity for Education
func (e *Education) GetID() int {
	return e.ID
}

// GetProfileID implements CRUDEntity for Education
func (e *Education) GetProfileID() int {
	return e.ProfileID
}

// GetID implements CRUDEntity for Certification
func (c *Certification) GetID() int {
	return c.ID
}

// GetProfileID implements CRUDEntity for Certification
func (c *Certification) GetProfileID() int {
	return c.ProfileID
}
