package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

const (
	serviceAuthHeader  = "X-Service-Api-Key"
	organizationHeader = "X-Organization-Id"
)

// NewAPIKeyMiddleware authenticates internal service callers via X-Service-Api-Key.
// The operator becomes SystemAppUserID and the tenant comes from X-Organization-Id.
// The header is deliberately distinct from public auth headers to prevent impersonation.
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
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("missing_api_key", "service API key is required"))
			c.Abort()
			return
		}

		if subtle.ConstantTimeCompare([]byte(key), []byte(expectedKey)) != 1 {
			logger.WarnContext(ctx, "invalid service API key")
			c.JSON(http.StatusForbidden, controller.NewErrorResponse("invalid_api_key", "service API key is invalid"))
			c.Abort()
			return
		}

		orgID := c.GetHeader(organizationHeader)
		if orgID == "" {
			logger.WarnContext(ctx, "missing organization id header")
			c.JSON(http.StatusBadRequest, controller.NewErrorResponse("missing_organization_id", "X-Organization-Id header is required"))
			c.Abort()
			return
		}

		// Populated context fields:
		//   - ContextFieldUserID:         SystemAppUserID (fixed bootstrap user)
		//   - ContextFieldOrganizationID: from X-Organization-Id header
		//
		// Intentionally omitted (not applicable to service-to-service calls):
		//   - ContextFieldOrganizationName
		//   - Any user-session fields set by the public auth middleware
		c.Set(controller.ContextFieldUserID{}, domain.SystemAppUserID)
		c.Set(controller.ContextFieldOrganizationID{}, orgID)
		c.Next()
	}
}
