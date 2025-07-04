package gemini

import (
	"errors"
	"fmt"
	"net/http"

	commonerrors "github.com/benidevo/vega/internal/common/errors"
)

var (
	// API Connection errors
	ErrAPIKeyInvalid      = commonerrors.New("invalid Gemini API key")
	ErrQuotaExceeded      = commonerrors.New("API quota exceeded")
	ErrServiceUnavailable = commonerrors.New("Gemini service temporarily unavailable")
	ErrRateLimitExceeded  = commonerrors.New("rate limit exceeded")
	ErrRequestTimeout     = commonerrors.New("request timeout")

	// Request/Response errors
	ErrInvalidRequest      = commonerrors.New("invalid request parameters")
	ErrInvalidResponse     = commonerrors.New("invalid response from Gemini")
	ErrEmptyResponse       = commonerrors.New("empty response from Gemini")
	ErrResponseParseFailed = commonerrors.New("failed to parse Gemini response")

	// Generation errors
	ErrCoverLetterGenFailed = commonerrors.New("cover letter generation failed")
	ErrCVGenFailed          = commonerrors.New("CV generation failed")
	ErrMatchAnalysisFailed  = commonerrors.New("job match analysis failed")

	// Technical/Infrastructure errors
	ErrClientInitFailed   = commonerrors.New("failed to initialize Gemini client")
	ErrMaxRetriesExceeded = commonerrors.New("maximum retry attempts exceeded")
)

// WrapError wraps the given innerErr with the provided sentinelErr.
// It returns a new error that combines both errors, preserving the original error context.
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}

// GetSentinelError returns the sentinel error corresponding to the provided error.
func GetSentinelError(err error) error {
	return commonerrors.GetSentinelError(err)
}

// GeminiError provides detailed API error information
type GeminiError struct {
	Code    int
	Message string
	Err     error
}

func (e *GeminiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("gemini API error (code %d): %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("gemini API error (code %d): %s", e.Code, e.Message)
}

func (e *GeminiError) Unwrap() error {
	return e.Err
}

// NewGeminiError creates a new instance of GeminiError with the specified code, message, and underlying error.
func NewGeminiError(code int, message string, err error) *GeminiError {
	return &GeminiError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsRetryableError determines whether the provided error is considered retryable.
//
// It checks if the error is a GeminiError with specific HTTP status codes (429, 500, 502, 503, 504)
// that typically warrant a retry. Also, it checks for sentinel errors such as
// ErrServiceUnavailable, ErrRateLimitExceeded, and ErrRequestTimeout, which also indicate
// retryable conditions. Returns true if the error is retryable, false otherwise.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var geminiErr *GeminiError
	if errors.As(err, &geminiErr) {
		switch geminiErr.Code {
		case http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		}
	}

	sentinelErr := GetSentinelError(err)
	return sentinelErr == ErrServiceUnavailable ||
		sentinelErr == ErrRateLimitExceeded ||
		sentinelErr == ErrRequestTimeout
}
