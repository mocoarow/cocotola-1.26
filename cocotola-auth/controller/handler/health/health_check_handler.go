// Package health provides the health check endpoint handler for the cocotola-auth service.
package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// Checker checks the health of the service.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckHandler handles the GET /auth/health endpoint.
type CheckHandler struct {
	checker Checker
	logger  *slog.Logger
}

// NewCheckHandler returns a new CheckHandler.
func NewCheckHandler(checker Checker) *CheckHandler {
	return &CheckHandler{
		checker: checker,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "HealthCheckHandler")),
	}
}

// HealthCheck handles GET /auth/health and verifies database connectivity.
func (h *CheckHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.checker.Check(ctx); err != nil {
		h.logger.ErrorContext(ctx, "health check", slog.Any("error", err))
		c.JSON(http.StatusServiceUnavailable, controller.NewErrorResponse("service_unavailable", "database is not available"))

		return
	}

	c.JSON(http.StatusOK, api.HealthCheckResponse{
		Status: "ok",
	})
}
