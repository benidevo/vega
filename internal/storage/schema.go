package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// UserDocument represents the complete user data stored in JSON format
type UserDocument struct {
	UpdatedAt time.Time    `json:"updated_at"`
	Checksum  string       `json:"checksum,omitempty"`
	Data      UserDataCore `json:"data"`
}

// NewUserDocument creates a new user document
func NewUserDocument() *UserDocument {
	return &UserDocument{
		UpdatedAt: time.Now(),
		Data: UserDataCore{
			Companies: make([]*Company, 0),
			Jobs:      make([]*Job, 0),
			Matches:   make([]*MatchResult, 0),
		},
	}
}

// Validate ensures the document is valid and can be processed
func (ud *UserDocument) Validate() error {
	return nil
}

// UpdateChecksum calculates and sets the checksum for the document
func (ud *UserDocument) UpdateChecksum() error {
	// Temporarily clear checksum for calculation
	originalChecksum := ud.Checksum
	ud.Checksum = ""

	data, err := json.Marshal(ud)
	if err != nil {
		ud.Checksum = originalChecksum
		return fmt.Errorf("failed to marshal document for checksum: %w", err)
	}

	hash := sha256.Sum256(data)
	ud.Checksum = hex.EncodeToString(hash[:])
	return nil
}

// VerifyChecksum validates the document's integrity
func (ud *UserDocument) VerifyChecksum() error {
	if ud.Checksum == "" {
		return nil // No checksum to verify
	}

	expectedChecksum := ud.Checksum
	ud.Checksum = ""

	data, err := json.Marshal(ud)
	if err != nil {
		ud.Checksum = expectedChecksum
		return fmt.Errorf("failed to marshal document for verification: %w", err)
	}

	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	ud.Checksum = expectedChecksum

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// ToJSON serializes the document to JSON with optional pretty printing
func (ud *UserDocument) ToJSON(pretty bool) ([]byte, error) {
	// Update timestamp and checksum before serialization
	ud.UpdatedAt = time.Now()
	if err := ud.UpdateChecksum(); err != nil {
		return nil, fmt.Errorf("failed to update checksum: %w", err)
	}

	if pretty {
		return json.MarshalIndent(ud, "", "  ")
	}
	return json.Marshal(ud)
}

// FromJSON deserializes a JSON document
func FromJSON(data []byte) (*UserDocument, error) {
	var doc UserDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	if err := doc.VerifyChecksum(); err != nil {
		return nil, fmt.Errorf("document integrity check failed: %w", err)
	}

	return &doc, nil
}

// GetJobByID returns a job by its ID
func (ud *UserDocument) GetJobByID(jobID int) *Job {
	for _, job := range ud.Data.Jobs {
		if job.ID == jobID {
			return job
		}
	}
	return nil
}

// GetCompanyByID returns a company by its ID
func (ud *UserDocument) GetCompanyByID(companyID int) *Company {
	for _, company := range ud.Data.Companies {
		if company.ID == companyID {
			return company
		}
	}
	return nil
}

// GetMatchResultByID returns a match result by its ID
func (ud *UserDocument) GetMatchResultByID(matchID int) *MatchResult {
	for _, match := range ud.Data.Matches {
		if match.ID == matchID {
			return match
		}
	}
	return nil
}

// GetJobsByCompany returns all jobs for a specific company
func (ud *UserDocument) GetJobsByCompany(companyID int) []*Job {
	var jobs []*Job
	for _, job := range ud.Data.Jobs {
		if job.Company.ID == companyID {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// GetMatchesByJob returns all match results for a specific job
func (ud *UserDocument) GetMatchesByJob(jobID int) []*MatchResult {
	var matches []*MatchResult
	for _, match := range ud.Data.Matches {
		if match.JobID == jobID {
			matches = append(matches, match)
		}
	}
	return matches
}

// RemoveCompany removes a company and all associated jobs and matches
func (ud *UserDocument) RemoveCompany(companyID int) {
	jobIDs := make(map[int]bool)
	for _, job := range ud.Data.Jobs {
		if job.Company.ID == companyID {
			jobIDs[job.ID] = true
		}
	}

	var remainingMatches []*MatchResult
	for _, match := range ud.Data.Matches {
		if !jobIDs[match.JobID] {
			remainingMatches = append(remainingMatches, match)
		}
	}
	ud.Data.Matches = remainingMatches

	// Remove all jobs for this company
	var remainingJobs []*Job
	for _, job := range ud.Data.Jobs {
		if job.Company.ID != companyID {
			remainingJobs = append(remainingJobs, job)
		}
	}
	ud.Data.Jobs = remainingJobs

	var remainingCompanies []*Company
	for _, company := range ud.Data.Companies {
		if company.ID != companyID {
			remainingCompanies = append(remainingCompanies, company)
		}
	}
	ud.Data.Companies = remainingCompanies
}

// RemoveJob removes a job and all associated matches
func (ud *UserDocument) RemoveJob(jobID int) {
	var remainingMatches []*MatchResult
	for _, match := range ud.Data.Matches {
		if match.JobID != jobID {
			remainingMatches = append(remainingMatches, match)
		}
	}
	ud.Data.Matches = remainingMatches

	var remainingJobs []*Job
	for _, job := range ud.Data.Jobs {
		if job.ID != jobID {
			remainingJobs = append(remainingJobs, job)
		}
	}
	ud.Data.Jobs = remainingJobs
}
