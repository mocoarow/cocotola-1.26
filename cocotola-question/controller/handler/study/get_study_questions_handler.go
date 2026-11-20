package study

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
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
	audio   *handler.AudioURLBuilder
	logger  *slog.Logger
}

// NewGetStudyQuestionsHandler returns a new GetStudyQuestionsHandler.
func NewGetStudyQuestionsHandler(usecase GetStudyQuestionsUsecase, audio *handler.AudioURLBuilder) *GetStudyQuestionsHandler {
	return &GetStudyQuestionsHandler{
		usecase: usecase,
		audio:   audio,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetStudyQuestionsHandler")),
	}
}

// GetStudyQuestions handles GET /workbook/:workbookId/study?limit=N.
func (h *GetStudyQuestionsHandler) GetStudyQuestions(c *gin.Context) {
	ctx := c.Request.Context()

	input, ok := h.parseStudyQuestionsInput(ctx, c)
	if !ok {
		return
	}

	output, err := h.usecase.GetStudyQuestions(ctx, input)
	if err != nil {
		controller.HandleError(ctx, h.logger, c, "get study questions", err)
		return
	}

	c.JSON(http.StatusOK, newGetStudyQuestionsResponse(output, h.audio))
}

func (h *GetStudyQuestionsHandler) parseStudyQuestionsInput(ctx context.Context, c *gin.Context) (*studyservice.GetStudyQuestionsInput, bool) {
	userID := c.GetString(controller.ContextFieldUserID{})
	if userID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return nil, false
	}

	organizationID := c.GetString(controller.ContextFieldOrganizationID{})
	if organizationID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid organization ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return nil, false
	}

	workbookID := c.Param("workbookId")
	if workbookID == "" {
		h.logger.WarnContext(ctx, "missing workbook ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "workbook ID is required"))
		return nil, false
	}

	limit, ok := h.parseLimit(ctx, c)
	if !ok {
		return nil, false
	}

	practice, ok := h.parsePractice(ctx, c)
	if !ok {
		return nil, false
	}

	input, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     userID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
		Limit:          limit,
		Practice:       practice,
		ExcludeIDs:     c.QueryArray("excludeIds"),
	})
	if err != nil {
		controller.HandleError(ctx, h.logger, c, "new get study questions input", err)
		return nil, false
	}
	return input, true
}

func (h *GetStudyQuestionsHandler) parseLimit(ctx context.Context, c *gin.Context) (int, bool) {
	limitStr := c.Query("limit")
	if limitStr == "" {
		h.logger.WarnContext(ctx, "missing limit parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "limit query parameter is required"))
		return 0, false
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid limit parameter", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "limit must be an integer"))
		return 0, false
	}
	return limit, true
}

func (h *GetStudyQuestionsHandler) parsePractice(ctx context.Context, c *gin.Context) (bool, bool) {
	practiceStr := c.Query("practice")
	if practiceStr == "" {
		return false, true
	}

	practice, err := strconv.ParseBool(practiceStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid practice parameter", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "practice must be a boolean"))
		return false, false
	}
	return practice, true
}

func newGetStudyQuestionsResponse(output *studyservice.GetStudyQuestionsOutput, audio *handler.AudioURLBuilder) api.GetStudyQuestionsResponse {
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
			Audio:        audio.QuestionAudio(q.Audio),
		})
	}

	return api.GetStudyQuestionsResponse{
		Questions:   questions,
		TotalDue:    int32(output.TotalDue),
		NewCount:    int32(output.NewCount),
		ReviewCount: int32(output.ReviewCount),
	}
}
