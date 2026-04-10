package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// CreateUserUsecase defines the use case method required by the CreateUserHandler.
type CreateUserUsecase interface {
	CreateAppUser(ctx context.Context, input *userservice.CreateAppUserInput) (*userservice.CreateAppUserOutput, error)
}

// CreateUserHandler handles the POST /user endpoint.
type CreateUserHandler struct {
	usecase CreateUserUsecase
	logger  *slog.Logger
}

// NewCreateUserHandler returns a new CreateUserHandler.
func NewCreateUserHandler(usecase CreateUserUsecase) *CreateUserHandler {
	return &CreateUserHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "CreateUserHandler")),
	}
}

// CreateUser handles POST /user.
func (h *CreateUserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationName := c.GetString(controller.ContextFieldOrganizationName{})
	if organizationName == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing organization name")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	var req api.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid create user request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	input, err := userservice.NewCreateAppUserInput(userID, organizationName, req.LoginId, req.Password)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid create user input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.CreateAppUser(ctx, input)
	if err != nil {
		h.handleError(ctx, c, err)
		return
	}

	c.JSON(http.StatusCreated, api.CreateUserResponse{
		AppUserID:      output.AppUserID.UUID(),
		OrganizationID: output.OrganizationID.UUID(),
		LoginID:        output.LoginID,
		Enabled:        output.Enabled,
	})
}

func (h *CreateUserHandler) handleError(ctx context.Context, c *gin.Context, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		h.logger.WarnContext(ctx, "forbidden", slog.Any("error", err))
		c.JSON(http.StatusForbidden, controller.NewErrorResponse("forbidden", http.StatusText(http.StatusForbidden)))
		return
	}
	if errors.Is(err, domain.ErrActiveUserLimitReached) || errors.Is(err, domain.ErrDuplicateEntry) {
		h.logger.WarnContext(ctx, "conflict", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", err.Error()))
		return
	}
	if errors.Is(err, domain.ErrOrganizationNotFound) {
		h.logger.WarnContext(ctx, "organization not found", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("organization_not_found", "organization not found"))
		return
	}
	h.logger.ErrorContext(ctx, "create user", slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
