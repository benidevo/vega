package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestRepositoryError_Error(t *testing.T) {
	sentinelErr := New("sentinel error")
	innerErr := fmt.Errorf("inner error")

	tests := []struct {
		name           string
		repositoryErr  *RepositoryError
		expectedOutput string
	}{
		{
			name: "With inner error",
			repositoryErr: &RepositoryError{
				SentinelError: sentinelErr,
				InnerError:    innerErr,
			},
			expectedOutput: "sentinel error: inner error",
		},
		{
			name: "Without inner error",
			repositoryErr: &RepositoryError{
				SentinelError: sentinelErr,
				InnerError:    nil,
			},
			expectedOutput: "sentinel error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repositoryErr.Error(); got != tt.expectedOutput {
				t.Errorf("RepositoryError.Error() = %v, want %v", got, tt.expectedOutput)
			}
		})
	}
}

func TestRepositoryError_Unwrap(t *testing.T) {
	sentinelErr := New("sentinel error")
	innerErr := fmt.Errorf("inner error")

	repoErr := &RepositoryError{
		SentinelError: sentinelErr,
		InnerError:    innerErr,
	}

	if unwrapped := repoErr.Unwrap(); unwrapped != sentinelErr {
		t.Errorf("RepositoryError.Unwrap() = %v, want %v", unwrapped, sentinelErr)
	}
}

func TestRepositoryError_Is(t *testing.T) {
	sentinelErr := New("sentinel error")
	otherErr := New("other error")
	innerErr := fmt.Errorf("inner error")

	repoErr := &RepositoryError{
		SentinelError: sentinelErr,
		InnerError:    innerErr,
	}

	tests := []struct {
		name       string
		target     error
		shouldBe   bool
		targetDesc string
	}{
		{
			name:       "Matching sentinel error",
			target:     sentinelErr,
			shouldBe:   true,
			targetDesc: "sentinel error",
		},
		{
			name:       "Non-matching error",
			target:     otherErr,
			shouldBe:   false,
			targetDesc: "other error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoErr.Is(tt.target); got != tt.shouldBe {
				t.Errorf("RepositoryError.Is(%s) = %v, want %v", tt.targetDesc, got, tt.shouldBe)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	sentinelErr := New("sentinel error")
	innerErr := fmt.Errorf("inner error")

	tests := []struct {
		name           string
		sentinelErr    error
		innerErr       error
		expectedResult error
		expectedType   interface{}
	}{
		{
			name:           "Wrap with inner error",
			sentinelErr:    sentinelErr,
			innerErr:       innerErr,
			expectedResult: &RepositoryError{SentinelError: sentinelErr, InnerError: innerErr},
			expectedType:   &RepositoryError{},
		},
		{
			name:           "Nil inner error",
			sentinelErr:    sentinelErr,
			innerErr:       nil,
			expectedResult: nil,
			expectedType:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.sentinelErr, tt.innerErr)

			if result == nil && tt.expectedResult == nil {
				return // Both are nil, test passes
			}

			if (result == nil && tt.expectedResult != nil) || (result != nil && tt.expectedResult == nil) {
				t.Errorf("WrapError() = %v, want %v", result, tt.expectedResult)
				return
			}

			if tt.expectedType != nil {
				var repoErr *RepositoryError
				if !stderrors.As(result, &repoErr) {
					t.Errorf("WrapError() result is not of type *RepositoryError")
				}
			}
		})
	}
}

func TestGetSentinelError(t *testing.T) {
	sentinelErr := New("sentinel error")
	innerErr := fmt.Errorf("inner error")
	repoErr := WrapError(sentinelErr, innerErr)
	plainErr := fmt.Errorf("plain error")

	tests := []struct {
		name         string
		err          error
		expectedErr  error
		expectedDesc string
	}{
		{
			name:         "Repository error",
			err:          repoErr,
			expectedErr:  sentinelErr,
			expectedDesc: "sentinel error",
		},
		{
			name:         "Plain error",
			err:          plainErr,
			expectedErr:  plainErr,
			expectedDesc: "plain error",
		},
		{
			name:         "Nil error",
			err:          nil,
			expectedErr:  nil,
			expectedDesc: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSentinelError(tt.err)

			if result == nil && tt.expectedErr == nil {
				return // Both are nil, test passes
			}

			if (result == nil && tt.expectedErr != nil) || (result != nil && tt.expectedErr == nil) {
				t.Errorf("GetSentinelError() = %v, want %v", result, tt.expectedErr)
				return
			}

			if result.Error() != tt.expectedErr.Error() {
				t.Errorf("GetSentinelError() = %v, want %v", result, tt.expectedDesc)
			}
		})
	}
}

func TestAs(t *testing.T) {
	sentinelErr := New("sentinel error")
	innerErr := fmt.Errorf("inner error")
	repoErr := WrapError(sentinelErr, innerErr)

	var testRepoErr *RepositoryError
	if !As(repoErr, &testRepoErr) {
		t.Errorf("As() failed to extract *RepositoryError")
	}

	if testRepoErr.SentinelError != sentinelErr {
		t.Errorf("Extracted error has wrong sentinel error: got %v, want %v",
			testRepoErr.SentinelError, sentinelErr)
	}

	if testRepoErr.InnerError != innerErr {
		t.Errorf("Extracted error has wrong inner error: got %v, want %v",
			testRepoErr.InnerError, innerErr)
	}

	// Use a different error type that won't match
	type customError struct{ error }
	var otherErr *customError
	if As(repoErr, &otherErr) {
		t.Errorf("As() incorrectly extracted a different error type")
	}
}

func TestIs(t *testing.T) {
	sentinelErr := New("sentinel error")
	otherErr := New("other error")
	repoErr := WrapError(sentinelErr, fmt.Errorf("inner error"))

	if !Is(repoErr, sentinelErr) {
		t.Errorf("Is() failed to match sentinel error")
	}

	if Is(repoErr, otherErr) {
		t.Errorf("Is() incorrectly matched a different error")
	}
}
