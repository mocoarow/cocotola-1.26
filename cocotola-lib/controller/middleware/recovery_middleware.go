package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// NewRecoveryMiddleware returns a Gin middleware that recovers from panics, records them on the
// request span, gin context errors and logger, and responds with 500.
func NewRecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		ctx := c.Request.Context()
		err := fmt.Errorf("panic: %v", recovered)

		trace.SpanFromContext(ctx).RecordError(err)
		_ = c.Error(err)

		logger.ErrorContext(ctx, "recovered from panic", "error", err, "stack", string(debug.Stack()))

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
