package study

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// RecordAnswerUsecase defines the use case method required by the RecordAnswerHandler.
type RecordAnswerUsecase interface {
	RecordAnswer(ctx context.Context, input *studyservice.RecordAnswerInput) (*studyservice.RecordAnswerOutput, error)
}

// RecordAnswerHandler handles the POST /workbook/:workbookId/study/:questionId/answer endpoint.
type RecordAnswerHandler struct {
	usecase RecordAnswerUsecase
	logger  *slog.Logger
}

// NewRecordAnswerHandler returns a new RecordAnswerHandler.
func NewRecordAnswerHandler(usecase RecordAnswerUsecase) *RecordAnswerHandler {
	return &RecordAnswerHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "RecordAnswerHandler")),
	}
}

// recordAnswerBody is a handler-level request struct that uses *bool to detect omission.
type recordAnswerBody struct {
	Correct *bool `json:"correct"`
}

// RecordAnswer handles POST /workbook/:workbookId/study/:questionId/answer.
func (h *RecordAnswerHandler) RecordAnswer(c *gin.Context) {
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

	questionID := c.Param("questionId")
	if questionID == "" {
		h.logger.WarnContext(ctx, "missing question ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "question ID is required"))
		return
	}

	var body recordAnswerBody
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		h.logger.WarnContext(ctx, "invalid record answer request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	if body.Correct == nil {
		h.logger.WarnContext(ctx, "missing correct field")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "correct field is required"))
		return
	}

	input, err := studyservice.NewRecordAnswerInput(userID, organizationID, workbookID, questionID, *body.Correct)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid record answer input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	output, err := h.usecase.RecordAnswer(ctx, input)
	if err != nil {
		handleStudyError(ctx, h.logger, c, "record answer", err)
		return
	}

	c.JSON(http.StatusOK, api.RecordAnswerResponse{
		NextDueAt:          output.NextDueAt,
		ConsecutiveCorrect: int32(output.ConsecutiveCorrect),
		TotalCorrect:       int32(output.TotalCorrect),
		TotalIncorrect:     int32(output.TotalIncorrect),
	})
}
