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

// PasswordAuthenticateUsecase defines the use case methods required by the PasswordAuthenticateHandler.
type PasswordAuthenticateUsecase interface {
	PasswordAuthenticate(ctx context.Context, input *authservice.PasswordAuthenticateInput) (*authservice.PasswordAuthenticateOutput, error)
	CreateSessionToken(ctx context.Context, input *authservice.CreateSessionTokenInput) (*authservice.CreateSessionTokenOutput, error)
	CreateTokenPair(ctx context.Context, input *authservice.CreateTokenPairInput) (*authservice.CreateTokenPairOutput, error)
}

// PasswordAuthenticateHandler handles the POST /auth/authenticate endpoint.
type PasswordAuthenticateHandler struct {
	usecase            PasswordAuthenticateUsecase
	cookieConfig       controller.CookieConfig
	sessionTokenTTLMin int
	logger             *slog.Logger
}

// NewPasswordAuthenticateHandler returns a new PasswordAuthenticateHandler.
func NewPasswordAuthenticateHandler(usecase PasswordAuthenticateUsecase, cookieConfig controller.CookieConfig, sessionTokenTTLMin int) *PasswordAuthenticateHandler {
	return &PasswordAuthenticateHandler{
		usecase:            usecase,
		cookieConfig:       cookieConfig,
		sessionTokenTTLMin: sessionTokenTTLMin,
		logger:             slog.Default().With(slog.String(liblogging.LoggerNameKey, "PasswordAuthenticateHandler")),
	}
}

// Authenticate handles POST /auth/authenticate.
// With X-Token-Delivery: cookie, creates a session token and sets it as a cookie.
// With X-Token-Delivery: json (or default), creates an access/refresh token pair and returns them in JSON.
func (h *PasswordAuthenticateHandler) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()
	var req api.PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid authenticate request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_authenticate_request", "request body is invalid"))
		return
	}

	tokenDelivery := c.GetHeader("X-Token-Delivery")
	switch tokenDelivery {
	case "", "json", "cookie":
		// valid
	default:
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_token_delivery", "X-Token-Delivery must be 'json' or 'cookie'"))
		return
	}

	input, err := authservice.NewPasswordAuthenticateInput(req.LoginID, req.Password, req.OrganizationName)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid authenticate input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	authOutput, err := h.usecase.PasswordAuthenticate(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthenticated) {
			h.logger.WarnContext(ctx, "unauthenticated", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthenticated", http.StatusText(http.StatusUnauthorized)))
			return
		}
		h.logger.ErrorContext(ctx, "authenticate", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	switch tokenDelivery {
	case "cookie":
		sessionInput, err := authservice.NewCreateSessionTokenInput(authOutput.UserID, authOutput.LoginID, authOutput.OrganizationName)
		if err != nil {
			h.logger.ErrorContext(ctx, "create session token input", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			return
		}

		sessionOutput, err := h.usecase.CreateSessionToken(ctx, sessionInput)
		if err != nil {
			h.logger.ErrorContext(ctx, "create session token", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			return
		}

		h.cookieConfig.SetTokenCookie(c.Writer, sessionOutput.RawToken, h.sessionTokenTTLMin)
		c.JSON(http.StatusOK, api.AuthenticateResponse{
			AccessToken:  nil,
			RefreshToken: nil,
		})
	default:
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
}
