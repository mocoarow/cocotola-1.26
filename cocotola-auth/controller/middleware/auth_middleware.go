package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libtelemetry "github.com/mocoarow/cocotola-1.26/cocotola-lib/telemetry"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// AuthUsecase defines the use case for validating tokens.
type AuthUsecase interface {
	ValidateSessionToken(ctx context.Context, input *authservice.ValidateSessionTokenInput) (*authservice.ValidateSessionTokenOutput, error)
	ExtendSessionToken(ctx context.Context, input *authservice.ExtendSessionTokenInput) error
	ValidateAccessToken(ctx context.Context, input *authservice.ValidateAccessTokenInput) (*authservice.ValidateAccessTokenOutput, error)
}

// NewAuthMiddleware returns a Gin middleware that validates Bearer tokens (JWT)
// or session cookies (opaque token) and sets the user identity in the Gin context.
func NewAuthMiddleware(authUsecase AuthUsecase, cookieConfig *controller.CookieConfig, sessionTokenTTLMin int) gin.HandlerFunc {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-AuthMiddleware"))

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctx, span := tracer.Start(ctx, "AuthMiddleware")
		defer span.End()

		// Try Bearer token first (JWT)
		if token := extractBearerToken(c); token != "" {
			output, err := authUsecase.ValidateAccessToken(ctx, &authservice.ValidateAccessTokenInput{JWTString: token})
			if err != nil {
				logger.WarnContext(ctx, "validate access token", slog.Any("error", err))
				c.Status(http.StatusUnauthorized)
				c.Abort()
				return
			}
			setUserContext(c, ctx, output.UserID, output.LoginID, output.OrganizationName, logger)
			c.Next()
			return
		}

		// Fall back to session cookie (opaque token)
		if cookieConfig != nil {
			cookie, err := c.Cookie(cookieConfig.Name)
			if err == nil && cookie != "" {
				output, err := authUsecase.ValidateSessionToken(ctx, &authservice.ValidateSessionTokenInput{RawToken: cookie})
				if err != nil {
					logger.WarnContext(ctx, "validate session token", slog.Any("error", err))
					c.Status(http.StatusUnauthorized)
					c.Abort()
					return
				}
				setUserContext(c, ctx, output.UserID, output.LoginID, output.OrganizationName, logger)

				// Sliding window: extend the session
				if err := authUsecase.ExtendSessionToken(ctx, &authservice.ExtendSessionTokenInput{RawToken: cookie}); err != nil {
					logger.WarnContext(ctx, "extend session token", slog.Any("error", err))
				} else {
					cookieConfig.SetTokenCookie(c.Writer, cookie, sessionTokenTTLMin)
				}

				c.Next()
				return
			}
		}

		logger.InfoContext(ctx, "no token found in Authorization header or cookie")
		c.Status(http.StatusUnauthorized)
		c.Abort()
	}
}

func extractBearerToken(c *gin.Context) string {
	authorization := c.GetHeader("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		return authorization[len("Bearer "):]
	}
	return ""
}

func setUserContext(c *gin.Context, ctx context.Context, userID int, loginID string, organizationName string, logger *slog.Logger) {
	c.Set(controller.ContextFieldUserID{}, userID)
	c.Set(controller.ContextFieldLoginID{}, loginID)
	c.Set(controller.ContextFieldOrganizationName{}, organizationName)
	ctx = libtelemetry.AddBaggageMembers(ctx, map[string]string{
		"user_id": strconv.Itoa(userID),
	}, logger)
	c.Request = c.Request.WithContext(ctx)
}
