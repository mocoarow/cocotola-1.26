package study

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// DeleteStudyHistoryUsecase defines the use case method required by DeleteStudyHistoryHandler.
type DeleteStudyHistoryUsecase interface {
	DeleteStudyHistory(ctx context.Context, input *studyservice.DeleteStudyHistoryInput) error
}

// DeleteStudyHistoryHandler handles DELETE /workbook/:workbookId/study.
type DeleteStudyHistoryHandler struct {
	usecase DeleteStudyHistoryUsecase
	logger  *slog.Logger
}

// NewDeleteStudyHistoryHandler returns a new DeleteStudyHistoryHandler.
func NewDeleteStudyHistoryHandler(usecase DeleteStudyHistoryUsecase) *DeleteStudyHistoryHandler {
	return &DeleteStudyHistoryHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "DeleteStudyHistoryHandler")),
	}
}

// DeleteStudyHistory handles DELETE /workbook/:workbookId/study.
func (h *DeleteStudyHistoryHandler) DeleteStudyHistory(c *gin.Context) {
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

	input, err := studyservice.NewDeleteStudyHistoryInput(userID, organizationID, workbookID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid delete study history input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	if err := h.usecase.DeleteStudyHistory(ctx, input); err != nil {
		handleStudyError(ctx, h.logger, c, "delete study history", err)
		return
	}

	c.Status(http.StatusNoContent)
}
