package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

const serviceAuthHeader = "X-Service-Api-Key"

// NewAPIKeyMiddleware returns a Gin middleware that validates the X-Service-Api-Key header
// using constant-time comparison for service-to-service authentication.
func NewAPIKeyMiddleware(expectedKey string) gin.HandlerFunc {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-APIKeyMiddleware"))

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctx, span := tracer.Start(ctx, "APIKeyMiddleware")
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		key := c.GetHeader(serviceAuthHeader)
		if key == "" {
			logger.WarnContext(ctx, "missing service API key")
			c.JSON(http.StatusUnauthorized, gin.H{"code": "missing_api_key", "message": "service API key is required"})
			c.Abort()

			return
		}

		if subtle.ConstantTimeCompare([]byte(key), []byte(expectedKey)) != 1 {
			logger.WarnContext(ctx, "invalid service API key")
			c.JSON(http.StatusForbidden, gin.H{"code": "invalid_api_key", "message": "service API key is invalid"})
			c.Abort()

			return
		}

		c.Next()
	}
}
