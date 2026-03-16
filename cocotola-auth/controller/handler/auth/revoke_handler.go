package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// RevokeUsecase defines the use case methods required by the RevokeHandler.
type RevokeUsecase interface {
	RevokeToken(ctx context.Context, input *authservice.RevokeTokenInput) error
	RevokeSessionToken(ctx context.Context, input *authservice.RevokeSessionTokenInput) error
}

// RevokeHandler handles the POST /auth/logout and POST /auth/revoke endpoints.
type RevokeHandler struct {
	usecase      RevokeUsecase
	logger       *slog.Logger
	cookieConfig *controller.CookieConfig
}

// NewRevokeHandler returns a new RevokeHandler.
func NewRevokeHandler(usecase RevokeUsecase, cookieConfig *controller.CookieConfig) *RevokeHandler {
	return &RevokeHandler{
		usecase:      usecase,
		logger:       slog.Default().With(slog.String(liblogging.LoggerNameKey, "RevokeHandler")),
		cookieConfig: cookieConfig,
	}
}

// Logout handles POST /auth/logout and clears the session cookie.
func (h *RevokeHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	if h.cookieConfig == nil {
		h.logger.ErrorContext(ctx, "logout requested but cookie config is not available")
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("cookie_not_configured", "cookie delivery is not configured"))
		return
	}

	// Try to revoke the session token if present
	cookie, err := c.Cookie(h.cookieConfig.Name)
	if err == nil && cookie != "" {
		revokeInput, err := authservice.NewRevokeSessionTokenInput(cookie)
		if err != nil {
			h.logger.WarnContext(ctx, "invalid session token on logout", slog.Any("error", err))
		} else if err := h.usecase.RevokeSessionToken(ctx, revokeInput); err != nil {
			h.logger.WarnContext(ctx, "revoke session token on logout", slog.Any("error", err))
		}
	}

	h.cookieConfig.ClearTokenCookie(c.Writer)
	c.Status(http.StatusNoContent)
}

// Revoke handles POST /auth/revoke and revokes the given token.
func (h *RevokeHandler) Revoke(c *gin.Context) {
	ctx := c.Request.Context()
	var req api.RevokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid revoke request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_revoke_request", "request body is invalid"))
		return
	}

	revokeInput, err := authservice.NewRevokeTokenInput(req.Token)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid revoke token input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err := h.usecase.RevokeToken(ctx, revokeInput); err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			h.logger.WarnContext(ctx, "token not found for revocation", slog.Any("error", err))
			c.JSON(http.StatusNotFound, controller.NewErrorResponse("token_not_found", "token not found"))
			return
		}
		h.logger.ErrorContext(ctx, "revoke token", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.Status(http.StatusNoContent)
}
