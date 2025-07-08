package storage

import "errors"

var (
	// ErrProfileNotFound is returned when a profile is not found
	ErrProfileNotFound = errors.New("profile not found")

	// ErrCompanyNotFound is returned when a company is not found
	ErrCompanyNotFound = errors.New("company not found")

	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.New("job not found")

	// ErrMatchNotFound is returned when a match result is not found
	ErrMatchNotFound = errors.New("match not found")

	// ErrNoMatches is returned when there are no matches
	ErrNoMatches = errors.New("no matches found")

	// ErrStorageNotInitialized is returned when storage is not initialized
	ErrStorageNotInitialized = errors.New("storage not initialized")
)
