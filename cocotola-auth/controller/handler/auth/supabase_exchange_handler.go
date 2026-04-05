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

// supabaseExchangeRequest is the request body for the Supabase token exchange endpoint.
type supabaseExchangeRequest struct {
	SupabaseJWT      string `json:"supabaseJwt" binding:"required"`
	OrganizationName string `json:"organizationName" binding:"required,max=20"`
}

// SupabaseExchangeUsecase defines the use case methods required by the SupabaseExchangeHandler.
type SupabaseExchangeUsecase interface {
	SupabaseExchange(ctx context.Context, input *authservice.SupabaseExchangeInput) (*authservice.SupabaseExchangeOutput, error)
	CreateTokenPair(ctx context.Context, input *authservice.CreateTokenPairInput) (*authservice.CreateTokenPairOutput, error)
}

// SupabaseExchangeHandler handles the POST /auth/supabase/exchange endpoint.
type SupabaseExchangeHandler struct {
	usecase SupabaseExchangeUsecase
	logger  *slog.Logger
}

// NewSupabaseExchangeHandler returns a new SupabaseExchangeHandler.
func NewSupabaseExchangeHandler(usecase SupabaseExchangeUsecase) *SupabaseExchangeHandler {
	return &SupabaseExchangeHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "SupabaseExchangeHandler")),
	}
}

// Exchange handles POST /auth/supabase/exchange.
func (h *SupabaseExchangeHandler) Exchange(c *gin.Context) {
	ctx := c.Request.Context()
	var req supabaseExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid supabase exchange request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_supabase_exchange_request", "request body is invalid"))
		return
	}

	input, err := authservice.NewSupabaseExchangeInput(req.SupabaseJWT, req.OrganizationName)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid supabase exchange input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	exchangeOutput, err := h.usecase.SupabaseExchange(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrOrganizationNotFound) {
			h.logger.WarnContext(ctx, "organization not found", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, controller.NewErrorResponse("organization_not_found", "organization not found"))
			return
		}
		h.logger.WarnContext(ctx, "supabase exchange", slog.Any("error", err))
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthenticated", "invalid supabase token"))
		return
	}

	tokenInput, err := authservice.NewCreateTokenPairInput(exchangeOutput.UserID, exchangeOutput.LoginID, exchangeOutput.OrganizationName)
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
