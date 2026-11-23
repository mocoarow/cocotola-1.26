package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// OrganizationIDResolver resolves an organization name to its ID.
type OrganizationIDResolver interface {
	ResolveOrganizationID(ctx context.Context, name string) (string, error)
}

// OrganizationIDResolverFunc adapts a function to OrganizationIDResolver.
type OrganizationIDResolverFunc func(ctx context.Context, name string) (string, error)

// ResolveOrganizationID calls f.
func (f OrganizationIDResolverFunc) ResolveOrganizationID(ctx context.Context, name string) (string, error) {
	return f(ctx, name)
}

// NewOrganizationResolverMiddleware resolves the organization name set by the
// auth middleware into the organization ID used by the handlers.
func NewOrganizationResolverMiddleware(resolver OrganizationIDResolver) gin.HandlerFunc {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-OrganizationResolverMiddleware"))

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		orgName := c.GetString(controller.ContextFieldOrganizationName{})
		if orgName == "" {
			logger.WarnContext(ctx, "missing organization name in context")
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
			c.Abort()
			return
		}

		orgID, err := resolver.ResolveOrganizationID(ctx, orgName)
		if err != nil {
			logger.ErrorContext(ctx, "resolve organization ID", slog.String("organization_name", orgName), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			c.Abort()
			return
		}

		c.Set(controller.ContextFieldOrganizationID{}, orgID)
		c.Next()
	}
}
