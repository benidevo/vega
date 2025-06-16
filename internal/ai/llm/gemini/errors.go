package gemini

import "errors"

var (
	ErrInvalidAPIKey   = errors.New("invalid Gemini API key")
	ErrQuotaExceeded   = errors.New("gemini API quota exceeded")
	ErrInvalidResponse = errors.New("invalid response from Gemini")
)
