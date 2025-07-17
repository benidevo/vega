package monitoring

// monitorConfig holds the configuration for the monitoring system
type monitorConfig struct {
	enabled     bool
	metricsPort string
	serviceName string
	version     string
	cloudMode   bool
	tokenSecret string
}
