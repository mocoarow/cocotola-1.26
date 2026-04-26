package workbook

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
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

	userID := c.GetString(controller.ContextFieldUserID{})
	if userID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationID := c.GetString(controller.ContextFieldOrganizationID{})
	if organizationID == "" {
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

	input, err := workbookservice.NewUpdateWorkbookInput(userID, organizationID, workbookID, req.Title, req.Description, string(req.Visibility), req.Language)
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

	c.JSON(http.StatusOK, api.WorkbookResponse{
		WorkbookID:     output.WorkbookID,
		SpaceID:        output.SpaceID,
		OwnerID:        output.OwnerID,
		OrganizationID: output.OrganizationID,
		Title:          output.Title,
		Description:    output.Description,
		Visibility:     api.WorkbookResponseVisibility(output.Visibility),
		Language:       output.Language,
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	})
}
