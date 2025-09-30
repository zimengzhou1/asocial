package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker is an interface for components that can be health checked
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	checker HealthChecker
	logger  *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(checker HealthChecker, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		checker: checker,
		logger:  logger,
	}
}

// HandleLiveness handles liveness probe endpoint
// Returns 200 OK if the application is running
func (h *HealthHandler) HandleLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleReadiness handles readiness probe endpoint
// Returns 200 OK if the application is ready to accept traffic (Redis is healthy)
func (h *HealthHandler) HandleReadiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)

	// Check Redis connection
	if err := h.checker.HealthCheck(ctx); err != nil {
		h.logger.Warn("Health check failed", "component", "redis", "error", err)
		checks["redis"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"checks": checks,
		})
		return
	}

	checks["redis"] = "ok"
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": checks,
	})
}
