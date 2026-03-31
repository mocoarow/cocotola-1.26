package workbook

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// DeleteWorkbookUsecase defines the use case method required by the DeleteWorkbookHandler.
type DeleteWorkbookUsecase interface {
	DeleteWorkbook(ctx context.Context, input *workbookservice.DeleteWorkbookInput) error
}

// DeleteWorkbookHandler handles the DELETE /workbook/:workbookId endpoint.
type DeleteWorkbookHandler struct {
	usecase DeleteWorkbookUsecase
	logger  *slog.Logger
}

// NewDeleteWorkbookHandler returns a new DeleteWorkbookHandler.
func NewDeleteWorkbookHandler(usecase DeleteWorkbookUsecase) *DeleteWorkbookHandler {
	return &DeleteWorkbookHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "DeleteWorkbookHandler")),
	}
}

// DeleteWorkbook handles DELETE /workbook/:workbookId.
func (h *DeleteWorkbookHandler) DeleteWorkbook(c *gin.Context) {
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

	input, err := workbookservice.NewDeleteWorkbookInput(userID, organizationID, workbookID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid delete workbook input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err := h.usecase.DeleteWorkbook(ctx, input); err != nil {
		handleWorkbookError(ctx, h.logger, c, "delete workbook", err)
		return
	}

	c.Status(http.StatusNoContent)
}
