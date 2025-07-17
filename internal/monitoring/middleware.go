package monitoring

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	RequestsTotal   metric.Int64Counter
	RequestDuration metric.Float64Histogram
	ActiveRequests  metric.Int64UpDownCounter
}

func (m *Monitor) CreateHTTPMetrics() (*HTTPMetrics, error) {
	requestsTotal, err := m.meter.Int64Counter(
		"vega_api_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := m.meter.Float64Histogram(
		"vega_api_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return nil, err
	}

	activeRequests, err := m.meter.Int64UpDownCounter(
		"vega_api_active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
		ActiveRequests:  activeRequests,
	}, nil
}

func (m *Monitor) GinMiddleware() gin.HandlerFunc {
	// First add OpenTelemetry middleware for tracing
	handlers := []gin.HandlerFunc{
		otelgin.Middleware(m.config.serviceName),
	}

	// Create HTTP metrics
	httpMetrics, err := m.CreateHTTPMetrics()
	if err != nil {
		m.log.Error().Err(err).Msg("Failed to create HTTP metrics")
		return func(c *gin.Context) { c.Next() }
	}

	// Add custom metrics middleware
	metricsMiddleware := func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Increment active requests
		httpMetrics.ActiveRequests.Add(c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("path", path),
			),
		)

		c.Next()

		// Decrement active requests
		httpMetrics.ActiveRequests.Add(c.Request.Context(), -1,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("path", path),
			),
		)

		// Record request metrics
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		attrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.String("path", path),
			attribute.String("status", status),
		}

		httpMetrics.RequestsTotal.Add(c.Request.Context(), 1,
			metric.WithAttributes(attrs...),
		)

		httpMetrics.RequestDuration.Record(c.Request.Context(), duration.Seconds(),
			metric.WithAttributes(attrs...),
		)
	}

	handlers = append(handlers, metricsMiddleware)

	// Return combined middleware
	return func(c *gin.Context) {
		for _, handler := range handlers {
			handler(c)
			if c.IsAborted() {
				return
			}
		}
	}
}
