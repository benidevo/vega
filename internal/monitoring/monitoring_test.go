package monitoring

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitoringSetup(t *testing.T) {
	settings := config.NewTestSettings()
	settings.MetricsEnabled = true
	settings.IsCloudMode = false

	monitor, err := Setup(&settings)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	ctx := context.Background()
	monitor.RecordQuotaUsage(ctx, "test-tenant", 50.0)
	monitor.RecordAIOperation(ctx, AIOpJobMatch, true, 100*time.Millisecond)

	// Wait for async metrics recording to complete
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	monitor.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, bodyStr, "vega_quota_usage_percentage")
	assert.Contains(t, bodyStr, "vega_ai_operations_total")
	assert.Contains(t, bodyStr, "vega_ai_operation_duration_seconds")

	err = monitor.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestMonitoringCloudMode(t *testing.T) {
	settings := config.NewTestSettings()
	settings.MetricsEnabled = true
	settings.IsCloudMode = true
	settings.TokenSecret = "test-secret"

	monitor, err := Setup(&settings)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	// Record some metrics to ensure they show up
	ctx := context.Background()
	monitor.RecordQuotaUsage(ctx, "test-tenant", 75.0)

	// Wait for async metrics recording to complete
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	monitor.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req = httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Authorization", "Bearer test-secret")
	w = httptest.NewRecorder()
	monitor.ServeHTTP(w, req)

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, bodyStr, "vega_quota_usage_percentage")

	err = monitor.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestAIOperationMetrics(t *testing.T) {
	settings := config.NewTestSettings()
	settings.MetricsEnabled = true
	settings.Version = "test"
	settings.IsCloudMode = false

	monitor, err := Setup(&settings)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	ctx := context.Background()

	recordComplete := monitor.RecordAIOperationStart(AIOpJobMatch)
	time.Sleep(50 * time.Millisecond)
	recordComplete(ctx, true)

	monitor.RecordAIOperation(ctx, AIOpCoverLetter, true, 200*time.Millisecond)
	monitor.RecordAIOperation(ctx, AIOpJobMatch, false, 500*time.Millisecond)

	// Wait for async metrics recording to complete
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	monitor.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check if metrics are present (they include a value after the labels)
	assert.Contains(t, bodyStr, "vega_ai_operations_total")
	assert.Contains(t, bodyStr, `operation="job_match"`)
	assert.Contains(t, bodyStr, `operation="cover_letter"`)
	assert.Contains(t, bodyStr, `success="true"`)
	assert.Contains(t, bodyStr, `success="false"`)

	assert.Contains(t, bodyStr, "vega_ai_operation_duration_seconds_bucket")

	err = monitor.Shutdown(ctx)
	assert.NoError(t, err)
}
