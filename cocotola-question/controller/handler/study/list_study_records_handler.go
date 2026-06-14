package study

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// ListStudyRecordsUsecase defines the use case method required by ListStudyRecordsHandler.
type ListStudyRecordsUsecase interface {
	ListStudyRecords(ctx context.Context, input *studyservice.ListStudyRecordsInput) (*studyservice.ListStudyRecordsOutput, error)
}

// ListStudyRecordsHandler handles GET /workbook/:workbookId/study/records.
type ListStudyRecordsHandler struct {
	usecase ListStudyRecordsUsecase
	logger  *slog.Logger
}

// NewListStudyRecordsHandler returns a new ListStudyRecordsHandler.
func NewListStudyRecordsHandler(usecase ListStudyRecordsUsecase) *ListStudyRecordsHandler {
	return &ListStudyRecordsHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListStudyRecordsHandler")),
	}
}

// ListStudyRecords handles GET /workbook/:workbookId/study/records.
func (h *ListStudyRecordsHandler) ListStudyRecords(c *gin.Context) {
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

	input, err := studyservice.NewListStudyRecordsInput(studyservice.ListStudyRecordsInputParams{
		OperatorID:     userID,
		OrganizationID: organizationID,
		WorkbookID:     workbookID,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list study records input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	output, err := h.usecase.ListStudyRecords(ctx, input)
	if err != nil {
		handleStudyError(ctx, h.logger, c, "list study records", err)
		return
	}

	records := make([]api.StudyRecordResponse, 0, len(output.Records))
	for _, r := range output.Records {
		records = append(records, api.StudyRecordResponse{
			WorkbookId:         r.WorkbookID,
			QuestionId:         r.QuestionID,
			ConsecutiveCorrect: int32(r.ConsecutiveCorrect),
			LastAnsweredAt:     r.LastAnsweredAt,
			NextDueAt:          r.NextDueAt,
			TotalCorrect:       int32(r.TotalCorrect),
			TotalIncorrect:     int32(r.TotalIncorrect),
		})
	}

	c.JSON(http.StatusOK, api.ListStudyRecordsResponse{Records: records})
}
