package health

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Handler handles health check requests
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new health check handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// HealthResponse represents the response structure for health checks
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Health `json:"checks,omitempty"`
}

// Health represents the health status of a component
type Health struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// GetHealth handles the /health endpoint for simple liveness check
func (h *Handler) GetHealth(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
	}

	log.Debug().Msg("Health check endpoint called")
	c.JSON(http.StatusOK, response)
}

// GetReady handles the /ready endpoint for readiness check with database
func (h *Handler) GetReady(c *gin.Context) {
	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]Health),
	}

	dbHealth := h.checkDatabase()
	response.Checks["database"] = dbHealth

	// Determine overall status
	if dbHealth.Status != "healthy" {
		response.Status = "not_ready"
		log.Warn().Str("db_status", dbHealth.Status).Msg("Readiness check failed")
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	log.Debug().Msg("Readiness check passed")
	c.JSON(http.StatusOK, response)
}

// checkDatabase verifies database connectivity and performance
func (h *Handler) checkDatabase() Health {
	if h.db == nil {
		return Health{
			Status:  "unhealthy",
			Message: "database connection is nil",
		}
	}

	start := time.Now()

	if err := h.db.Ping(); err != nil {
		return Health{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: time.Since(start),
		}
	}

	latency := time.Since(start)

	if latency > 1*time.Second {
		return Health{
			Status:  "degraded",
			Message: "database response time is slow",
			Latency: latency,
		}
	}

	return Health{
		Status:  "healthy",
		Message: "database is responding",
		Latency: latency,
	}
}
