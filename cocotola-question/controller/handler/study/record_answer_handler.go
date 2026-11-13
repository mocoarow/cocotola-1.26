package study

import (
	"context"
	"encoding/json"
	"fmt"
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

// recordAnswerBody is the loose handler-level request struct. Pointer fields
// detect omission; per-type guards (right field for the question's type) live
// in the usecase, since the handler does not yet know the question type.
type recordAnswerBody struct {
	Correct           *bool       `json:"correct"`
	SelectedChoiceIDs *[]string   `json:"selectedChoiceIds"`
	Regions           *[][]string `json:"regions"`
}

func (b *recordAnswerBody) answerFieldCount() int {
	count := 0
	if b.Correct != nil {
		count++
	}
	if b.SelectedChoiceIDs != nil {
		count++
	}
	if b.Regions != nil {
		count++
	}
	return count
}

// RecordAnswer handles POST /workbook/:workbookId/study/:questionId/answer.
func (h *RecordAnswerHandler) RecordAnswer(c *gin.Context) {
	ctx := c.Request.Context()

	ids, ok := h.parseAnswerIDs(ctx, c)
	if !ok {
		return
	}

	body, ok := h.parseAnswerBody(ctx, c)
	if !ok {
		return
	}

	input, ok := h.buildAnswerInput(ctx, c, ids, body)
	if !ok {
		return
	}

	output, err := h.usecase.RecordAnswer(ctx, input)
	if err != nil {
		controller.HandleError(ctx, h.logger, c, "record answer", err)
		return
	}

	c.JSON(http.StatusOK, api.RecordAnswerResponse{
		NextDueAt:          output.NextDueAt,
		ConsecutiveCorrect: int32(output.ConsecutiveCorrect),
		TotalCorrect:       int32(output.TotalCorrect),
		TotalIncorrect:     int32(output.TotalIncorrect),
	})
}

// answerIDs holds the user/organization/workbook/question identifiers parsed from the request.
type answerIDs struct {
	userID         string
	organizationID string
	workbookID     string
	questionID     string
}

func (h *RecordAnswerHandler) parseAnswerIDs(ctx context.Context, c *gin.Context) (*answerIDs, bool) {
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

	questionID := c.Param("questionId")
	if questionID == "" {
		h.logger.WarnContext(ctx, "missing question ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "question ID is required"))
		return nil, false
	}

	return &answerIDs{userID: userID, organizationID: organizationID, workbookID: workbookID, questionID: questionID}, true
}

func (h *RecordAnswerHandler) parseAnswerBody(ctx context.Context, c *gin.Context) (*recordAnswerBody, bool) {
	var body recordAnswerBody
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		h.logger.WarnContext(ctx, "invalid record answer request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return nil, false
	}
	return &body, true
}

func (h *RecordAnswerHandler) buildAnswerInput(ctx context.Context, c *gin.Context, ids *answerIDs, body *recordAnswerBody) (*studyservice.RecordAnswerInput, bool) {
	count := body.answerFieldCount()
	if count == 0 {
		h.logger.WarnContext(ctx, "missing answer field")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "one of correct, selectedChoiceIds or regions must be provided"))
		return nil, false
	}
	if count > 1 {
		h.logger.WarnContext(ctx, "ambiguous answer field")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "correct, selectedChoiceIds and regions are mutually exclusive"))
		return nil, false
	}

	in, err := newAnswerInput(ids, body, c.GetHeader("X-Local-Date"), c.GetHeader("X-Local-Timezone"))
	if err != nil {
		controller.HandleError(ctx, h.logger, c, "new record answer input", err)
		return nil, false
	}
	return in, true
}

// newAnswerInput expects exactly one answer field in body to be set.
func newAnswerInput(ids *answerIDs, body *recordAnswerBody, localDateKey, timezone string) (*studyservice.RecordAnswerInput, error) {
	switch {
	case body.Correct != nil:
		in, err := studyservice.NewRecordAnswerInputForWordFill(studyservice.RecordAnswerInputForWordFillParams{
			OperatorID:     ids.userID,
			OrganizationID: ids.organizationID,
			WorkbookID:     ids.workbookID,
			QuestionID:     ids.questionID,
			Correct:        *body.Correct,
			LocalDateKey:   localDateKey,
			Timezone:       timezone,
		})
		if err != nil {
			return nil, fmt.Errorf("new word_fill answer input: %w", err)
		}
		return in, nil
	case body.SelectedChoiceIDs != nil:
		in, err := studyservice.NewRecordAnswerInputForMultipleChoice(studyservice.RecordAnswerInputForMultipleChoiceParams{
			OperatorID:        ids.userID,
			OrganizationID:    ids.organizationID,
			WorkbookID:        ids.workbookID,
			QuestionID:        ids.questionID,
			SelectedChoiceIDs: *body.SelectedChoiceIDs,
			LocalDateKey:      localDateKey,
			Timezone:          timezone,
		})
		if err != nil {
			return nil, fmt.Errorf("new multiple_choice answer input: %w", err)
		}
		return in, nil
	default:
		in, err := studyservice.NewRecordAnswerInputForOrdering(studyservice.RecordAnswerInputForOrderingParams{
			OperatorID:     ids.userID,
			OrganizationID: ids.organizationID,
			WorkbookID:     ids.workbookID,
			QuestionID:     ids.questionID,
			Regions:        *body.Regions,
			LocalDateKey:   localDateKey,
			Timezone:       timezone,
		})
		if err != nil {
			return nil, fmt.Errorf("new ordering answer input: %w", err)
		}
		return in, nil
	}
}
