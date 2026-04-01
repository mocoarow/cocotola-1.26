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

// GetQuestionUsecase defines the use case method required by the GetQuestionHandler.
type GetQuestionUsecase interface {
	GetQuestion(ctx context.Context, input *questionservice.GetQuestionInput) (*questionservice.GetQuestionOutput, error)
}

// GetQuestionHandler handles the GET /workbook/:workbookId/question/:questionId endpoint.
type GetQuestionHandler struct {
	usecase GetQuestionUsecase
	logger  *slog.Logger
}

// NewGetQuestionHandler returns a new GetQuestionHandler.
func NewGetQuestionHandler(usecase GetQuestionUsecase) *GetQuestionHandler {
	return &GetQuestionHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetQuestionHandler")),
	}
}

// GetQuestion handles GET /workbook/:workbookId/question/:questionId.
func (h *GetQuestionHandler) GetQuestion(c *gin.Context) {
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

	input, err := questionservice.NewGetQuestionInput(userID, organizationID, workbookID, questionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid get question input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.GetQuestion(ctx, input)
	if err != nil {
		handleQuestionError(ctx, h.logger, c, "get question", err)
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
