package study

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// GetStudyQuestionsUsecase defines the use case method required by the GetStudyQuestionsHandler.
type GetStudyQuestionsUsecase interface {
	GetStudyQuestions(ctx context.Context, input *studyservice.GetStudyQuestionsInput) (*studyservice.GetStudyQuestionsOutput, error)
}

// GetStudyQuestionsHandler handles the GET /workbook/:workbookId/study endpoint.
type GetStudyQuestionsHandler struct {
	usecase GetStudyQuestionsUsecase
	logger  *slog.Logger
}

// NewGetStudyQuestionsHandler returns a new GetStudyQuestionsHandler.
func NewGetStudyQuestionsHandler(usecase GetStudyQuestionsUsecase) *GetStudyQuestionsHandler {
	return &GetStudyQuestionsHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetStudyQuestionsHandler")),
	}
}

// GetStudyQuestions handles GET /workbook/:workbookId/study?limit=N.
func (h *GetStudyQuestionsHandler) GetStudyQuestions(c *gin.Context) {
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

	limitStr := c.Query("limit")
	if limitStr == "" {
		h.logger.WarnContext(ctx, "missing limit parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "limit query parameter is required"))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid limit parameter", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "limit must be an integer"))
		return
	}

	practice := false
	if practiceStr := c.Query("practice"); practiceStr != "" {
		v, err := strconv.ParseBool(practiceStr)
		if err != nil {
			h.logger.WarnContext(ctx, "invalid practice parameter", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "practice must be a boolean"))
			return
		}
		practice = v
	}

	input, err := studyservice.NewGetStudyQuestionsInput(userID, organizationID, workbookID, limit, practice)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid get study questions input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	output, err := h.usecase.GetStudyQuestions(ctx, input)
	if err != nil {
		handleStudyError(ctx, h.logger, c, "get study questions", err)
		return
	}

	questions := make([]api.StudyQuestionResponse, 0, len(output.Questions))
	for _, q := range output.Questions {
		tags := q.Tags
		if tags == nil {
			tags = []string{}
		}
		questions = append(questions, api.StudyQuestionResponse{
			QuestionId:   q.QuestionID,
			QuestionType: q.QuestionType,
			Content:      q.Content,
			Tags:         tags,
			OrderIndex:   int32(q.OrderIndex),
		})
	}

	c.JSON(http.StatusOK, api.GetStudyQuestionsResponse{
		Questions:   questions,
		TotalDue:    int32(output.TotalDue),
		NewCount:    int32(output.NewCount),
		ReviewCount: int32(output.ReviewCount),
	})
}
