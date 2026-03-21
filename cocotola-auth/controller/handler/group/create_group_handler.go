package group

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	groupservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/group"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// CreateGroupUsecase defines the use case method required by the CreateGroupHandler.
type CreateGroupUsecase interface {
	CreateGroup(ctx context.Context, input *groupservice.CreateGroupInput) (*groupservice.CreateGroupOutput, error)
}

// CreateGroupHandler handles the POST /group endpoint.
type CreateGroupHandler struct {
	usecase CreateGroupUsecase
	logger  *slog.Logger
}

// NewCreateGroupHandler returns a new CreateGroupHandler.
func NewCreateGroupHandler(usecase CreateGroupUsecase) *CreateGroupHandler {
	return &CreateGroupHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "CreateGroupHandler")),
	}
}

// CreateGroup handles POST /group.
func (h *CreateGroupHandler) CreateGroup(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(controller.ContextFieldUserID{})
	if userID <= 0 {
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

	var req api.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid create group request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	input, err := groupservice.NewCreateGroupInput(userID, organizationName, req.Name)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid create group input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.CreateGroup(ctx, input)
	if err != nil {
		h.handleError(ctx, c, err)
		return
	}

	groupID, err := safeIntToInt32(output.GroupID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert group ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	orgID, err := safeIntToInt32(output.OrganizationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert organization ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusCreated, api.CreateGroupResponse{
		GroupID:        groupID,
		OrganizationID: orgID,
		Name:           output.Name,
		Enabled:        output.Enabled,
	})
}

func (h *CreateGroupHandler) handleError(ctx context.Context, c *gin.Context, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		h.logger.WarnContext(ctx, "forbidden", slog.Any("error", err))
		c.JSON(http.StatusForbidden, controller.NewErrorResponse("forbidden", http.StatusText(http.StatusForbidden)))
		return
	}
	if errors.Is(err, domain.ErrActiveGroupLimitReached) || errors.Is(err, domain.ErrDuplicateEntry) {
		h.logger.WarnContext(ctx, "conflict", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", err.Error()))
		return
	}
	if errors.Is(err, domain.ErrOrganizationNotFound) {
		h.logger.WarnContext(ctx, "organization not found", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("organization_not_found", "organization not found"))
		return
	}
	h.logger.ErrorContext(ctx, "create group", slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}

func safeIntToInt32(v int) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d overflows int32", v)
	}
	return int32(v), nil
}
