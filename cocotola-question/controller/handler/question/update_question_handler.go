package question

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// UpdateQuestionUsecase defines the use case method required by the UpdateQuestionHandler.
type UpdateQuestionUsecase interface {
	UpdateQuestion(ctx context.Context, input *questionservice.UpdateQuestionInput) (*questionservice.UpdateQuestionOutput, error)
}

// UpdateQuestionHandler handles the PUT /workbook/:workbookId/question/:questionId endpoint.
type UpdateQuestionHandler struct {
	usecase UpdateQuestionUsecase
	logger  *slog.Logger
}

// NewUpdateQuestionHandler returns a new UpdateQuestionHandler.
func NewUpdateQuestionHandler(usecase UpdateQuestionUsecase) *UpdateQuestionHandler {
	return &UpdateQuestionHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "UpdateQuestionHandler")),
	}
}

// UpdateQuestion handles PUT /workbook/:workbookId/question/:questionId.
func (h *UpdateQuestionHandler) UpdateQuestion(c *gin.Context) {
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

	var req api.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid update question request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	input, err := questionservice.NewUpdateQuestionInput(userID, organizationID, workbookID, questionID, req.Content, int(req.OrderIndex))
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid update question input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.UpdateQuestion(ctx, input)
	if err != nil {
		handleQuestionError(ctx, h.logger, c, "update question", err)
		return
	}

	c.JSON(http.StatusOK, api.QuestionResponse{
		QuestionID:   output.QuestionID,
		QuestionType: output.QuestionType,
		Content:      output.Content,
		OrderIndex:   int32(output.OrderIndex),
		CreatedAt:    output.CreatedAt,
		UpdatedAt:    output.UpdatedAt,
	})
}
