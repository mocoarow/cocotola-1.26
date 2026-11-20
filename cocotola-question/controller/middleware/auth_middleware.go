package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// AuthMeFetcher resolves the caller identified by an Authorization header value.
type AuthMeFetcher interface {
	FetchMe(ctx context.Context, authorization string) (domain.AuthenticatedUser, error)
}

// NewAuthMiddleware authenticates end-user requests by resolving the Authorization header through fetcher.
func NewAuthMiddleware(fetcher AuthMeFetcher) gin.HandlerFunc {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-AuthMiddleware"))

	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "AuthMiddleware")
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			logger.InfoContext(ctx, "no Authorization header")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		me, err := fetcher.FetchMe(ctx, authorization)
		if err != nil {
			c.AbortWithStatus(authErrorStatus(ctx, logger, err))
			return
		}

		c.Set(controller.ContextFieldUserID{}, me.UserID)
		c.Set(controller.ContextFieldOrganizationName{}, me.OrganizationName)
		c.Set(controller.ContextFieldUserLanguage{}, me.Language)
		c.Next()
	}
}

func authErrorStatus(ctx context.Context, logger *slog.Logger, err error) int {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		logger.InfoContext(ctx, "auth service rejected credentials")
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrAuthServiceUnavailable):
		logger.ErrorContext(ctx, "auth service unavailable", slog.Any("error", err))
		return http.StatusBadGateway
	default:
		logger.ErrorContext(ctx, "fetch auth me", slog.Any("error", err))
		return http.StatusInternalServerError
	}
}
