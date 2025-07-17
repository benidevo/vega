package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/benidevo/vega/internal/common/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
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

	quotaUsage          metric.Float64Gauge
	aiOperations        metric.Int64Counter
	aiOperationDuration metric.Float64Histogram
	httpMetrics         *HTTPMetrics

	eventChan   chan metricEvent
	workerCount int
	wg          sync.WaitGroup
	shutdownCh  chan struct{}
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

	// Determine worker count and channel size based on environment
	workerCount := 2     // Conservative default for resource-constrained environments
	channelSize := 10000 // Larger buffer to handle bursts without dropping events

	m := &Monitor{
		config:         cfg,
		meter:          meter,
		meterProvider:  provider,
		metricsHandler: promhttp.Handler(),
		log:            logger.GetLogger("monitoring"),
		eventChan:      make(chan metricEvent, channelSize),
		workerCount:    workerCount,
		shutdownCh:     make(chan struct{}),
	}

	if err := m.createMetrics(); err != nil {
		return nil, fmt.Errorf("failed to create metrics: %w", err)
	}

	m.startWorkers()

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

	if err := m.createHTTPMetrics(); err != nil {
		return fmt.Errorf("failed to create HTTP metrics: %w", err)
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

// startWorkers starts the metric processing workers
func (m *Monitor) startWorkers() {
	for i := 0; i < m.workerCount; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
	m.log.Info().Int("workers", m.workerCount).Msg("Started metric workers")
}

// worker processes metric events from the channel
func (m *Monitor) worker(id int) {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.eventChan:
			m.processEvent(event)
		case <-m.shutdownCh:
			m.log.Debug().Int("worker_id", id).Msg("Worker shutting down")
			return
		}
	}
}

// RecordQuotaUsage records the current quota usage for a tenant
func (m *Monitor) RecordQuotaUsage(ctx context.Context, tenantID string, usage float64) {
	if m == nil {
		return
	}

	event := newQuotaUsageEvent(ctx, tenantID, usage)

	// Send event to channel (will block if full)
	select {
	case m.eventChan <- event:
		// Event sent successfully
	case <-ctx.Done():
		// Context cancelled, abandon metric
		m.log.Debug().Str("event_type", "quota_usage").Msg("Context cancelled, abandoning metric")
	}
}

// RecordAIOperation records an AI operation with its duration
func (m *Monitor) RecordAIOperation(ctx context.Context, operation string, success bool, duration time.Duration) {
	if m == nil {
		return
	}

	event := newAIOperationEvent(ctx, operation, success, duration)

	// Send event to channel (will block if full)
	select {
	case m.eventChan <- event:
		// Event sent successfully
	case <-ctx.Done():
		// Context cancelled, abandon metric
		m.log.Debug().Str("event_type", "ai_operation").Str("operation", operation).Msg("Context cancelled, abandoning metric")
	}
}

// Shutdown gracefully shuts down the monitoring system
func (m *Monitor) Shutdown(ctx context.Context) error {
	if m == nil {
		return nil
	}

	close(m.shutdownCh)

	// Wait for workers to finish processing with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.log.Info().Msg("All metric workers stopped")
	case <-ctx.Done():
		m.log.Warn().Msg("Shutdown timeout reached, some metrics may be lost")
	}

	close(m.eventChan)

	if m.meterProvider != nil {
		return m.meterProvider.Shutdown(ctx)
	}
	return nil
}
