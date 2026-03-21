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

// GuestAuthenticateUsecase defines the use case methods required by the GuestAuthenticateHandler.
type GuestAuthenticateUsecase interface {
	GuestAuthenticate(ctx context.Context, input *authservice.GuestAuthenticateInput) (*authservice.GuestAuthenticateOutput, error)
	CreateTokenPair(ctx context.Context, input *authservice.CreateTokenPairInput) (*authservice.CreateTokenPairOutput, error)
}

// GuestAuthenticateHandler handles the POST /auth/guest/authenticate endpoint.
type GuestAuthenticateHandler struct {
	usecase GuestAuthenticateUsecase
	logger  *slog.Logger
}

// NewGuestAuthenticateHandler returns a new GuestAuthenticateHandler.
func NewGuestAuthenticateHandler(usecase GuestAuthenticateUsecase) *GuestAuthenticateHandler {
	return &GuestAuthenticateHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GuestAuthenticateHandler")),
	}
}

// Authenticate handles POST /auth/guest/authenticate.
func (h *GuestAuthenticateHandler) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()
	var req api.GuestAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid guest authenticate request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_guest_authenticate_request", "request body is invalid"))
		return
	}

	input, err := authservice.NewGuestAuthenticateInput(req.OrganizationName)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid guest authenticate input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	authOutput, err := h.usecase.GuestAuthenticate(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthenticated) {
			h.logger.WarnContext(ctx, "guest unauthenticated", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthenticated", http.StatusText(http.StatusUnauthorized)))
			return
		}
		h.logger.ErrorContext(ctx, "guest authenticate", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	tokenInput, err := authservice.NewCreateTokenPairInput(authOutput.UserID, authOutput.LoginID, authOutput.OrganizationName)
	if err != nil {
		h.logger.ErrorContext(ctx, "create token pair input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	tokenOutput, err := h.usecase.CreateTokenPair(ctx, tokenInput)
	if err != nil {
		h.logger.ErrorContext(ctx, "create token pair", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	resp := api.AuthenticateResponse{
		AccessToken:  &tokenOutput.AccessToken,
		RefreshToken: &tokenOutput.RefreshToken,
	}
	c.JSON(http.StatusOK, resp)
}
