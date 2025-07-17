package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/config"
)

func BenchmarkChannelBasedMetrics(b *testing.B) {
	settings := config.NewTestSettings()
	settings.MetricsEnabled = true
	settings.IsCloudMode = true
	settings.TokenSecret = "test-secret"

	monitor, err := Setup(&settings)
	if err != nil {
		b.Fatal(err)
	}
	defer monitor.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordAIOperation(ctx, AIOpJobMatch, true, 100*time.Millisecond)
		}
	})

	// Allow workers to process remaining events
	time.Sleep(100 * time.Millisecond)
}

func BenchmarkQuotaUsageRecording(b *testing.B) {
	settings := config.NewTestSettings()
	settings.MetricsEnabled = true
	settings.IsCloudMode = true

	monitor, err := Setup(&settings)
	if err != nil {
		b.Fatal(err)
	}
	defer monitor.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordQuotaUsage(ctx, "tenant-123", 75.5)
	}
}
