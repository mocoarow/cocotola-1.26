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

// ListQuestionsUsecase defines the use case method required by the ListQuestionsHandler.
type ListQuestionsUsecase interface {
	ListQuestions(ctx context.Context, input *questionservice.ListQuestionsInput) (*questionservice.ListQuestionsOutput, error)
}

// ListQuestionsHandler handles the GET /workbook/:workbookId/question endpoint.
type ListQuestionsHandler struct {
	usecase ListQuestionsUsecase
	logger  *slog.Logger
}

// NewListQuestionsHandler returns a new ListQuestionsHandler.
func NewListQuestionsHandler(usecase ListQuestionsUsecase) *ListQuestionsHandler {
	return &ListQuestionsHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListQuestionsHandler")),
	}
}

// ListQuestions handles GET /workbook/:workbookId/question.
func (h *ListQuestionsHandler) ListQuestions(c *gin.Context) {
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

	input, err := questionservice.NewListQuestionsInput(userID, organizationID, workbookID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid list questions input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ListQuestions(ctx, input)
	if err != nil {
		handleQuestionError(ctx, h.logger, c, "list questions", err)
		return
	}

	items := make([]api.QuestionResponse, len(output.Questions))
	for i, q := range output.Questions {
		items[i] = api.QuestionResponse{
			QuestionID:   q.QuestionID,
			QuestionType: q.QuestionType,
			Content:      q.Content,
			Tags:         q.Tags,
			OrderIndex:   int32(q.OrderIndex),
			CreatedAt:    q.CreatedAt,
			UpdatedAt:    q.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, api.ListQuestionsResponse{
		Questions: items,
	})
}
