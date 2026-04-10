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

// CreateWorkbookUsecase defines the use case method required by the CreateWorkbookHandler.
type CreateWorkbookUsecase interface {
	CreateWorkbook(ctx context.Context, input *workbookservice.CreateWorkbookInput) (*workbookservice.CreateWorkbookOutput, error)
}

// CreateWorkbookHandler handles the POST /workbook endpoint.
type CreateWorkbookHandler struct {
	usecase CreateWorkbookUsecase
	logger  *slog.Logger
}

// NewCreateWorkbookHandler returns a new CreateWorkbookHandler.
func NewCreateWorkbookHandler(usecase CreateWorkbookUsecase) *CreateWorkbookHandler {
	return &CreateWorkbookHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "CreateWorkbookHandler")),
	}
}

// CreateWorkbook handles POST /workbook.
func (h *CreateWorkbookHandler) CreateWorkbook(c *gin.Context) {
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

	var req api.CreateWorkbookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid create workbook request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	input, err := workbookservice.NewCreateWorkbookInput(userID, organizationID, int(req.SpaceID), req.Title, description, string(req.Visibility))
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid create workbook input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.CreateWorkbook(ctx, input)
	if err != nil {
		handleWorkbookError(ctx, h.logger, c, "create workbook", err)
		return
	}

	spaceID, err := handler.SafeIntToInt32(output.SpaceID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert space ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusCreated, api.WorkbookResponse{
		WorkbookID:     output.WorkbookID,
		SpaceID:        spaceID,
		OwnerID:        output.OwnerID,
		OrganizationID: output.OrganizationID,
		Title:          output.Title,
		Description:    output.Description,
		Visibility:     api.WorkbookResponseVisibility(output.Visibility),
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
	})
}
