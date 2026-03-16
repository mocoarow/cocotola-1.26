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

// RefreshUsecase defines the use case methods required by the RefreshHandler.
type RefreshUsecase interface {
	RefreshAccessToken(ctx context.Context, input *authservice.RefreshAccessTokenInput) (*authservice.RefreshAccessTokenOutput, error)
}

// RefreshHandler handles the POST /auth/refresh endpoint.
type RefreshHandler struct {
	usecase RefreshUsecase
	logger  *slog.Logger
}

// NewRefreshHandler returns a new RefreshHandler.
func NewRefreshHandler(usecase RefreshUsecase) *RefreshHandler {
	return &RefreshHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "RefreshHandler")),
	}
}

// Refresh handles POST /auth/refresh and returns a new access token.
func (h *RefreshHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	var req api.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid refresh request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_refresh_request", "request body is invalid"))
		return
	}

	refreshInput, err := authservice.NewRefreshAccessTokenInput(req.RefreshToken)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid refresh input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	refreshOutput, err := h.usecase.RefreshAccessToken(ctx, refreshInput)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) || errors.Is(err, domain.ErrTokenRevoked) || errors.Is(err, domain.ErrSessionExpired) {
			h.logger.WarnContext(ctx, "refresh token invalid", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("invalid_refresh_token", "refresh token is invalid or expired"))
			return
		}
		h.logger.ErrorContext(ctx, "refresh access token", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.RefreshResponse{
		AccessToken: refreshOutput.AccessToken,
	})
}
