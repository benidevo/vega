package monitoring

import (
	"context"
	"time"
)

// AIOperation represents different AI operations we track
const (
	AIOpJobMatch      = "job_match"
	AIOpCVParse       = "cv_parse"
	AIOpCoverLetter   = "cover_letter"
	AIOpResumeBuilder = "resume_builder"
)

// RecordAIOperationStart returns a function to call when the operation completes
func (m *Monitor) RecordAIOperationStart(operation string) func(ctx context.Context, success bool) {
	if m == nil {
		return func(ctx context.Context, success bool) {}
	}

	start := time.Now()
	return func(ctx context.Context, success bool) {
		duration := time.Since(start)
		m.RecordAIOperation(ctx, operation, success, duration)
	}
}

// AIServiceMiddleware provides a middleware pattern for AI operations
func (m *Monitor) AIServiceMiddleware(operation string, fn func() error) error {
	if m == nil {
		return fn()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	success := err == nil
	m.RecordAIOperation(context.Background(), operation, success, duration)

	return err
}

// AIServiceMiddlewareWithContext provides a middleware pattern for AI operations with context
func (m *Monitor) AIServiceMiddlewareWithContext(ctx context.Context, operation string, fn func(context.Context) error) error {
	if m == nil {
		return fn(ctx)
	}

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	success := err == nil
	m.RecordAIOperation(ctx, operation, success, duration)

	return err
}
