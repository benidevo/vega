package job

import (
	"strings"

	"github.com/benidevo/vega/internal/job/models"
	"github.com/gin-gonic/gin"
)

// FieldCommand defines the interface for field update commands
type FieldCommand interface {
	Execute(c *gin.Context, job *models.Job, service *JobService) (string, error)
}

// statusCommand handles status field updates
type statusCommand struct{}

// Execute updates the status of the given job based on the "status" form value from the request context.
func (cmd *statusCommand) Execute(c *gin.Context, job *models.Job, service *JobService) (string, error) {
	statusStr := c.PostForm("status")
	if statusStr == "" {
		return "", models.ErrStatusRequired
	}

	status, err := models.JobStatusFromString(statusStr)
	if err != nil {
		return "", models.ErrInvalidJobStatus
	}

	job.Status = status
	return "Job status updated to " + status.String(), nil
}

// notesCommand handles notes field updates
type notesCommand struct{}

// Execute updates the Notes field of the given job with the value from the POST form and returns a success message.
func (cmd *notesCommand) Execute(c *gin.Context, job *models.Job, service *JobService) (string, error) {
	notes := strings.TrimSpace(c.PostForm("notes"))
	job.Notes = notes
	return "Notes updated successfully", nil
}

// skillsCommand handles skills field updates
type skillsCommand struct{}

// Execute processes the "skills" form field, validates and filters the skills,
// updates the job's RequiredSkills, and returns a success message or error.
func (cmd *skillsCommand) Execute(c *gin.Context, job *models.Job, service *JobService) (string, error) {
	skillsStr := c.PostForm("skills")
	skills := service.ValidateAndFilterSkills(skillsStr)
	job.RequiredSkills = skills
	return "Skills updated successfully", nil
}

// basicCommand handles basic job information updates
type basicCommand struct{}

// Execute updates the job fields (title, company name, and location) from the POST form data in the Gin context.
func (cmd *basicCommand) Execute(c *gin.Context, job *models.Job, service *JobService) (string, error) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		return "", models.ErrJobTitleRequired
	}
	job.Title = title

	companyName := strings.TrimSpace(c.PostForm("company_name"))
	if companyName == "" {
		return "", models.ErrCompanyNameRequired
	}
	job.Company.Name = companyName

	location := strings.TrimSpace(c.PostForm("location"))
	job.Location = location

	return "Job details updated successfully", nil
}

// CommandFactory creates field commands based on the field name
type CommandFactory struct {
	commands map[string]FieldCommand
}

// NewCommandFactory creates a new command factory
func NewCommandFactory() *CommandFactory {
	return &CommandFactory{
		commands: map[string]FieldCommand{
			"status": &statusCommand{},
			"notes":  &notesCommand{},
			"skills": &skillsCommand{},
			"basic":  &basicCommand{},
		},
	}
}

// GetCommand returns the command for the given field
func (f *CommandFactory) GetCommand(field string) (FieldCommand, error) {
	cmd, exists := f.commands[field]
	if !exists {
		return nil, models.ErrInvalidFieldParam
	}
	return cmd, nil
}
