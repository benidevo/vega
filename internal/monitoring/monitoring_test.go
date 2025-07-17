package monitoring

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	monitor.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.True(t, strings.Contains(bodyStr, `vega_ai_operations_total{operation="job_match",success="true"}`))
	assert.True(t, strings.Contains(bodyStr, `vega_ai_operations_total{operation="job_match",success="false"}`))
	assert.True(t, strings.Contains(bodyStr, `vega_ai_operations_total{operation="cover_letter",success="true"}`))

	assert.Contains(t, bodyStr, "vega_ai_operation_duration_seconds_bucket")

	err = monitor.Shutdown(ctx)
	assert.NoError(t, err)
}
