package models

import "time"

// GeneratedCV represents a CV generated for a specific job application
type GeneratedCV struct {
	JobID          int              `json:"jobId"`
	UserID         int              `json:"userId"`
	IsValid        bool             `json:"isValid"`
	Reason         string           `json:"reason,omitempty"`
	PersonalInfo   PersonalInfo     `json:"personalInfo,omitempty"`
	WorkExperience []WorkExperience `json:"workExperience,omitempty"`
	Education      []Education      `json:"education,omitempty"`
	Skills         []string         `json:"skills,omitempty"`
	GeneratedAt    time.Time        `json:"generatedAt"`
	JobTitle       string           `json:"jobTitle"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// PersonalInfo contains basic personal information for a CV
type PersonalInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Location  string `json:"location,omitempty"`
	Title     string `json:"title,omitempty"`
	Summary   string `json:"summary,omitempty"`
}

// WorkExperience represents a work experience entry in a CV
type WorkExperience struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	Location    string `json:"location,omitempty"`
	StartDate   string `json:"startDate"`         // Format: "YYYY-MM" or "YYYY"
	EndDate     string `json:"endDate,omitempty"` // Format: "YYYY-MM" or "YYYY" or "Present"
	Description string `json:"description,omitempty"`
}

// Education represents an education entry in a CV
type Education struct {
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy,omitempty"`
	StartDate    string `json:"startDate"`         // Format: "YYYY-MM" or "YYYY"
	EndDate      string `json:"endDate,omitempty"` // Format: "YYYY-MM" or "YYYY"
}
