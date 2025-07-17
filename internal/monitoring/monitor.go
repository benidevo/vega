package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/benidevo/vega/internal/common/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Monitor provides monitoring capabilities for the application
type Monitor struct {
	config         monitorConfig
	meter          metric.Meter
	meterProvider  *sdkmetric.MeterProvider
	metricsHandler http.Handler
	log            zerolog.Logger

	// Custom metrics
	quotaUsage          metric.Float64Gauge
	aiOperations        metric.Int64Counter
	aiOperationDuration metric.Float64Histogram
}

// new creates a new Monitor instance
func new(cfg monitorConfig) (*Monitor, error) {
	if !cfg.enabled {
		return nil, nil
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.serviceName),
			semconv.ServiceVersion(cfg.version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(provider)

	meter := provider.Meter(cfg.serviceName)

	m := &Monitor{
		config:         cfg,
		meter:          meter,
		meterProvider:  provider,
		metricsHandler: promhttp.Handler(),
		log:            logger.GetLogger("monitoring"),
	}

	if err := m.createMetrics(); err != nil {
		return nil, fmt.Errorf("failed to create metrics: %w", err)
	}

	return m, nil
}

func (m *Monitor) createMetrics() error {
	var err error

	m.quotaUsage, err = m.meter.Float64Gauge(
		"vega_quota_usage_percentage",
		metric.WithDescription("Current quota usage as percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return fmt.Errorf("failed to create quota usage metric: %w", err)
	}

	m.aiOperations, err = m.meter.Int64Counter(
		"vega_ai_operations_total",
		metric.WithDescription("Total number of AI operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return fmt.Errorf("failed to create AI operations metric: %w", err)
	}

	m.aiOperationDuration, err = m.meter.Float64Histogram(
		"vega_ai_operation_duration_seconds",
		metric.WithDescription("AI operation duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2.5, 5, 10, 30),
	)
	if err != nil {
		return fmt.Errorf("failed to create AI operation duration metric: %w", err)
	}

	return nil
}

// ServeHTTP serves the metrics endpoint
func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.config.cloudMode {
		token := r.Header.Get("Authorization")
		expectedToken := "Bearer " + m.config.tokenSecret
		if token != expectedToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	m.metricsHandler.ServeHTTP(w, r)
}

// RecordQuotaUsage records the current quota usage for a tenant
func (m *Monitor) RecordQuotaUsage(ctx context.Context, tenantID string, usage float64) {
	m.quotaUsage.Record(ctx, usage,
		metric.WithAttributes(
			attribute.String("tenant_id", tenantID),
		),
	)
}

// RecordAIOperation records an AI operation with its duration
func (m *Monitor) RecordAIOperation(ctx context.Context, operation string, success bool, duration time.Duration) {
	successStr := "false"
	if success {
		successStr = "true"
	}

	m.aiOperations.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("success", successStr),
		),
	)

	m.aiOperationDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("success", successStr),
		),
	)
}

// Shutdown gracefully shuts down the monitoring system
func (m *Monitor) Shutdown(ctx context.Context) error {
	if m.meterProvider != nil {
		return m.meterProvider.Shutdown(ctx)
	}
	return nil
}
