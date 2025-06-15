package health

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers health check routes
func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	handler := NewHandler(db)

	router.GET("/health", handler.GetHealth)

	router.GET("/ready", handler.GetReady)
}
