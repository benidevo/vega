package models

// CreateJobRequest represents the request payload for creating a job
type CreateJobRequest struct {
	Title          string `json:"title" binding:"required"`
	Company        string `json:"company" binding:"required"`
	Location       string `json:"location" binding:"required"`
	Description    string `json:"description" binding:"required"`
	JobType        string `json:"jobType,omitempty"`
	ApplicationURL string `json:"applicationUrl,omitempty"`
	SourceURL      string `json:"sourceUrl" binding:"required"`
	Notes          string `json:"notes,omitempty"`
}

// CreateJobResponse represents the response after creating a job
type CreateJobResponse struct {
	Message string `json:"message"`
	JobID   int    `json:"jobId,omitempty"`
}
