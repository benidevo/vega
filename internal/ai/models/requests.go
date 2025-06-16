package models

// Prompt represents the structure for a prompt used in the application.
type Prompt struct {
	Instructions     string `json:"instructions"`
	ApplicantName    string `json:"applicant_name"`
	ApplicantProfile string `json:"applicant_profile"`
	JobDescription   string `json:"job_description"`
	ExtraContext     string `json:"extra_context,omitempty"`
}
