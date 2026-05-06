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

// GetStudySummaryUsecase defines the use case method required by GetStudySummaryHandler.
type GetStudySummaryUsecase interface {
	GetStudySummary(ctx context.Context, input *studyservice.GetStudySummaryInput) (*studyservice.GetStudySummaryOutput, error)
}

// GetStudySummaryHandler handles GET /workbook/:workbookId/study/summary.
type GetStudySummaryHandler struct {
	usecase GetStudySummaryUsecase
	logger  *slog.Logger
}

// NewGetStudySummaryHandler returns a new GetStudySummaryHandler.
func NewGetStudySummaryHandler(usecase GetStudySummaryUsecase) *GetStudySummaryHandler {
	return &GetStudySummaryHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetStudySummaryHandler")),
	}
}

// GetStudySummary handles GET /workbook/:workbookId/study/summary.
func (h *GetStudySummaryHandler) GetStudySummary(c *gin.Context) {
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

	input, err := studyservice.NewGetStudySummaryInput(userID, organizationID, workbookID, practice)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid get study summary input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	output, err := h.usecase.GetStudySummary(ctx, input)
	if err != nil {
		handleStudyError(ctx, h.logger, c, "get study summary", err)
		return
	}

	c.JSON(http.StatusOK, api.StudySummaryResponse{
		NewCount:               int32(output.NewCount),
		ReviewCount:            int32(output.ReviewCount),
		TotalDue:               int32(output.TotalDue),
		ReviewRatioNumerator:   int32(output.ReviewRatioNumerator),
		ReviewRatioDenominator: int32(output.ReviewRatioDenominator),
	})
}
