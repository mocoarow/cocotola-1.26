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

// AddQuestionUsecase defines the use case method required by the AddQuestionHandler.
type AddQuestionUsecase interface {
	AddQuestion(ctx context.Context, input *questionservice.AddQuestionInput) (*questionservice.AddQuestionOutput, error)
}

// AddQuestionHandler handles the POST /workbook/:workbookId/question endpoint.
type AddQuestionHandler struct {
	usecase AddQuestionUsecase
	logger  *slog.Logger
}

// NewAddQuestionHandler returns a new AddQuestionHandler.
func NewAddQuestionHandler(usecase AddQuestionUsecase) *AddQuestionHandler {
	return &AddQuestionHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "AddQuestionHandler")),
	}
}

// AddQuestion handles POST /workbook/:workbookId/question.
func (h *AddQuestionHandler) AddQuestion(c *gin.Context) {
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

	var req api.AddQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid add question request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	input, err := questionservice.NewAddQuestionInput(userID, organizationID, workbookID, req.QuestionType, req.Content, int(req.OrderIndex))
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid add question input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.AddQuestion(ctx, input)
	if err != nil {
		handleQuestionError(ctx, h.logger, c, "add question", err)
		return
	}

	c.JSON(http.StatusCreated, api.QuestionResponse{
		QuestionID:   output.QuestionID,
		QuestionType: output.QuestionType,
		Content:      output.Content,
		OrderIndex:   int32(output.OrderIndex),
		CreatedAt:    output.CreatedAt,
		UpdatedAt:    output.UpdatedAt,
	})
}
