package models

// MatchResult represents the result of a matching process, including the match score,
// identified strengths and weaknesses, key highlights, and overall feedback.
type MatchResult struct {
	MatchScore int      `json:"matchScore"`
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
	Highlights []string `json:"highlights"`
	Feedback   string   `json:"feedback"`
}

// CoverLetterFormat defines the format type for a cover letter, such as HTML, Markdown, or plain text.
// It is used to specify how the cover letter content is structured and presented.
type CoverLetterFormat string

const (
	CoverLetterTypeHtml      CoverLetterFormat = "html"
	CoverLetterTypeMarkdown  CoverLetterFormat = "markdown"
	CoverLetterTypePlainText CoverLetterFormat = "plain_text"
)

// CoverLetter represents a cover letter with its format and content.
type CoverLetter struct {
	Format  CoverLetterFormat `json:"format"`
	Content string            `json:"content"`
}

// CVParsingResult represents the structured data extracted from a CV/resume
type CVParsingResult struct {
	IsValid        bool             `json:"isValid"`
	Reason         string           `json:"reason,omitempty"`
	PersonalInfo   PersonalInfo     `json:"personalInfo,omitempty"`
	WorkExperience []WorkExperience `json:"workExperience,omitempty"`
	Education      []Education      `json:"education,omitempty"`
	Skills         []string         `json:"skills,omitempty"`
}

// PersonalInfo contains basic personal information from a CV
type PersonalInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Location  string `json:"location,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Title     string `json:"title,omitempty"`
	Summary   string `json:"summary,omitempty"`
}

// WorkExperience represents a work experience entry from a CV
type WorkExperience struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	Location    string `json:"location,omitempty"`
	StartDate   string `json:"startDate"`         // Format: "YYYY-MM" or "YYYY"
	EndDate     string `json:"endDate,omitempty"` // Format: "YYYY-MM" or "YYYY" or "Present"
	Description string `json:"description,omitempty"`
}

// Education represents an education entry from a CV
type Education struct {
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy,omitempty"`
	StartDate    string `json:"startDate"`         // Format: "YYYY-MM" or "YYYY"
	EndDate      string `json:"endDate,omitempty"` // Format: "YYYY-MM" or "YYYY"
}

// GeneratedCV represents a CV generated for a specific job application
type GeneratedCV struct {
	CVParsingResult
	GeneratedAt int64  `json:"generatedAt"` // Unix timestamp
	JobID       int    `json:"jobId"`
	JobTitle    string `json:"jobTitle"`
}
