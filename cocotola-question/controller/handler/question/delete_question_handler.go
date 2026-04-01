package question

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// DeleteQuestionUsecase defines the use case method required by the DeleteQuestionHandler.
type DeleteQuestionUsecase interface {
	DeleteQuestion(ctx context.Context, input *questionservice.DeleteQuestionInput) error
}

// DeleteQuestionHandler handles the DELETE /workbook/:workbookId/question/:questionId endpoint.
type DeleteQuestionHandler struct {
	usecase DeleteQuestionUsecase
	logger  *slog.Logger
}

// NewDeleteQuestionHandler returns a new DeleteQuestionHandler.
func NewDeleteQuestionHandler(usecase DeleteQuestionUsecase) *DeleteQuestionHandler {
	return &DeleteQuestionHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "DeleteQuestionHandler")),
	}
}

// DeleteQuestion handles DELETE /workbook/:workbookId/question/:questionId.
func (h *DeleteQuestionHandler) DeleteQuestion(c *gin.Context) {
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

	questionID := c.Param("questionId")
	if questionID == "" {
		h.logger.WarnContext(ctx, "missing question ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "question ID is required"))
		return
	}

	input, err := questionservice.NewDeleteQuestionInput(userID, organizationID, workbookID, questionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid delete question input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err := h.usecase.DeleteQuestion(ctx, input); err != nil {
		handleQuestionError(ctx, h.logger, c, "delete question", err)
		return
	}

	c.Status(http.StatusNoContent)
}
