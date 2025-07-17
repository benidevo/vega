package monitoring

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// eventType represents different types of metric events
type eventType int

const (
	eventTypeQuotaUsage eventType = iota
	eventTypeAIOperation
	eventTypeAIOperationDuration
	eventTypeHTTPRequest
	eventTypeHTTPDuration
)

// metricEvent represents a metric recording event
type metricEvent struct {
	eventType eventType
	ctx       context.Context
	timestamp time.Time

	// Common fields
	attrs []attribute.KeyValue

	// Event-specific fields
	floatValue float64
	intValue   int64
	duration   time.Duration
}

// newQuotaUsageEvent creates a quota usage metric event
func newQuotaUsageEvent(ctx context.Context, tenantID string, usage float64) metricEvent {
	return metricEvent{
		eventType:  eventTypeQuotaUsage,
		ctx:        ctx,
		timestamp:  time.Now(),
		floatValue: usage,
		attrs: []attribute.KeyValue{
			attribute.String("tenant_id", tenantID),
		},
	}
}

// newAIOperationEvent creates an AI operation metric event
func newAIOperationEvent(ctx context.Context, operation string, success bool, duration time.Duration) metricEvent {
	successStr := "false"
	if success {
		successStr = "true"
	}

	return metricEvent{
		eventType: eventTypeAIOperation,
		ctx:       ctx,
		timestamp: time.Now(),
		duration:  duration,
		intValue:  1,
		attrs: []attribute.KeyValue{
			attribute.String("operation", operation),
			attribute.String("success", successStr),
		},
	}
}

// newHTTPRequestEvent creates an HTTP request metric event
func newHTTPRequestEvent(ctx context.Context, method, path, status string, duration time.Duration) metricEvent {
	return metricEvent{
		eventType: eventTypeHTTPRequest,
		ctx:       ctx,
		timestamp: time.Now(),
		duration:  duration,
		intValue:  1,
		attrs: []attribute.KeyValue{
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.String("status", status),
		},
	}
}

// processEvent processes a metric event using the appropriate metric instrument
func (m *Monitor) processEvent(event metricEvent) {
	recordCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch event.eventType {
	case eventTypeQuotaUsage:
		m.quotaUsage.Record(recordCtx, event.floatValue,
			metric.WithAttributes(event.attrs...),
		)

	case eventTypeAIOperation:
		m.aiOperations.Add(recordCtx, event.intValue,
			metric.WithAttributes(event.attrs...),
		)
		m.aiOperationDuration.Record(recordCtx, event.duration.Seconds(),
			metric.WithAttributes(event.attrs...),
		)

	case eventTypeHTTPRequest:
		if httpMetrics := m.httpMetrics; httpMetrics != nil {
			httpMetrics.RequestsTotal.Add(recordCtx, event.intValue,
				metric.WithAttributes(event.attrs...),
			)
			httpMetrics.RequestDuration.Record(recordCtx, event.duration.Seconds(),
				metric.WithAttributes(event.attrs...),
			)
		}
	}
}
