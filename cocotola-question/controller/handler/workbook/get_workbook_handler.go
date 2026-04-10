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

// GetWorkbookUsecase defines the use case method required by the GetWorkbookHandler.
type GetWorkbookUsecase interface {
	GetWorkbook(ctx context.Context, input *workbookservice.GetWorkbookInput) (*workbookservice.GetWorkbookOutput, error)
}

// GetWorkbookHandler handles the GET /workbook/:workbookId endpoint.
type GetWorkbookHandler struct {
	usecase GetWorkbookUsecase
	logger  *slog.Logger
}

// NewGetWorkbookHandler returns a new GetWorkbookHandler.
func NewGetWorkbookHandler(usecase GetWorkbookUsecase) *GetWorkbookHandler {
	return &GetWorkbookHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetWorkbookHandler")),
	}
}

// GetWorkbook handles GET /workbook/:workbookId.
func (h *GetWorkbookHandler) GetWorkbook(c *gin.Context) {
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

	input, err := workbookservice.NewGetWorkbookInput(userID, organizationID, workbookID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid get workbook input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.GetWorkbook(ctx, input)
	if err != nil {
		handleWorkbookError(ctx, h.logger, c, "get workbook", err)
		return
	}

	spaceID, err := handler.SafeIntToInt32(output.SpaceID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert space ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.WorkbookResponse{
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
