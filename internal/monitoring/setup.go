package monitoring

import (
	"errors"

	"github.com/benidevo/vega/internal/config"
)

// Setup initializes the monitoring system with the provided settings
func Setup(settings *config.Settings) (*Monitor, error) {
	if !settings.MetricsEnabled {
		return nil, nil
	}

	if settings.IsCloudMode && settings.TokenSecret == "" {
		return nil, errors.New("TOKEN_SECRET is required for metrics in cloud mode")
	}

	cfg := monitorConfig{
		enabled:     settings.MetricsEnabled,
		metricsPort: settings.MetricsPort,
		serviceName: "vega-ai",
		version:     settings.Version,
		cloudMode:   settings.IsCloudMode,
		tokenSecret: settings.TokenSecret,
	}

	if cfg.metricsPort == "" {
		cfg.metricsPort = "9090"
	}

	if cfg.version == "" {
		cfg.version = "dev"
	}

	return new(cfg)
}
