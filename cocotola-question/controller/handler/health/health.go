// Package health provides the health check endpoint handler for the cocotola-question service.
package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// Checker checks the health of the service.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckHandler handles the GET /question/health endpoint.
type CheckHandler struct {
	checker Checker
	logger  *slog.Logger
}

// NewCheckHandler returns a new CheckHandler.
func NewCheckHandler(checker Checker) *CheckHandler {
	return &CheckHandler{
		checker: checker,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "QuestionHealthCheckHandler")),
	}
}

// HealthCheck handles GET /question/health and verifies Firestore connectivity.
func (h *CheckHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.checker.Check(ctx); err != nil {
		h.logger.ErrorContext(ctx, "health check", slog.Any("error", err))
		c.JSON(http.StatusServiceUnavailable, controller.NewErrorResponse("service_unavailable", "datastore is not available"))

		return
	}

	c.JSON(http.StatusOK, api.HealthCheckResponse{
		Status: "ok",
	})
}

// InitRouter sets up the routes for health operations under the given parent router group.
func InitRouter(
	checkHandler *CheckHandler,
	parentRouterGroup gin.IRouter,
	middleware ...gin.HandlerFunc,
) {
	questionGroup := parentRouterGroup.Group("question", middleware...)

	questionGroup.GET("/health", checkHandler.HealthCheck)
}
