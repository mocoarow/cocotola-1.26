package workbook

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// UpdateWorkbookUsecase defines the use case method required by the UpdateWorkbookHandler.
type UpdateWorkbookUsecase interface {
	UpdateWorkbook(ctx context.Context, input *workbookservice.UpdateWorkbookInput) (*workbookservice.UpdateWorkbookOutput, error)
}

// UpdateWorkbookHandler handles the PUT /workbook/:workbookId endpoint.
type UpdateWorkbookHandler struct {
	usecase UpdateWorkbookUsecase
	logger  *slog.Logger
}

// NewUpdateWorkbookHandler returns a new UpdateWorkbookHandler.
func NewUpdateWorkbookHandler(usecase UpdateWorkbookUsecase) *UpdateWorkbookHandler {
	return &UpdateWorkbookHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "UpdateWorkbookHandler")),
	}
}

// UpdateWorkbook handles PUT /workbook/:workbookId.
func (h *UpdateWorkbookHandler) UpdateWorkbook(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(controller.ContextFieldUserID{})
	if userID <= 0 {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationID := c.GetInt(controller.ContextFieldOrganizationID{})
	if organizationID <= 0 {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid organization ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	workbookID := c.Param("workbookId")
	if workbookID == "" {
		h.logger.WarnContext(ctx, "missing workbook ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "workbook ID is required"))
		return
	}

	var req api.UpdateWorkbookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid update workbook request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	input, err := workbookservice.NewUpdateWorkbookInput(userID, organizationID, workbookID, req.Title, description, string(req.Visibility))
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid update workbook input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.UpdateWorkbook(ctx, input)
	if err != nil {
		handleWorkbookError(ctx, h.logger, c, "update workbook", err)
		return
	}

	spaceID, err := handler.SafeIntToInt32(output.SpaceID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert space ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	ownerID, err := handler.SafeIntToInt32(output.OwnerID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert owner ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	orgID, err := handler.SafeIntToInt32(output.OrganizationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert organization ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.WorkbookResponse{
		WorkbookID:     output.WorkbookID,
		SpaceID:        spaceID,
		OwnerID:        ownerID,
		OrganizationID: orgID,
		Title:          output.Title,
		Description:    output.Description,
		Visibility:     api.WorkbookResponseVisibility(output.Visibility),
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	})
}
