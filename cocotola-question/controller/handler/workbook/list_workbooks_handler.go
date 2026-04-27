package workbook

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// ListWorkbooksUsecase defines the use case method required by the ListWorkbooksHandler.
type ListWorkbooksUsecase interface {
	ListWorkbooks(ctx context.Context, input *workbookservice.ListWorkbooksInput) (*workbookservice.ListWorkbooksOutput, error)
}

// ListWorkbooksHandler handles the GET /workbook endpoint.
type ListWorkbooksHandler struct {
	usecase ListWorkbooksUsecase
	logger  *slog.Logger
}

// NewListWorkbooksHandler returns a new ListWorkbooksHandler.
func NewListWorkbooksHandler(usecase ListWorkbooksUsecase) *ListWorkbooksHandler {
	return &ListWorkbooksHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListWorkbooksHandler")),
	}
}

// ListWorkbooks handles GET /workbook?spaceId=123.
func (h *ListWorkbooksHandler) ListWorkbooks(c *gin.Context) {
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

	spaceID := c.Query("spaceId")
	if spaceID == "" {
		h.logger.WarnContext(ctx, "missing spaceId query parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "spaceId query parameter is required"))
		return
	}

	input, err := workbookservice.NewListWorkbooksInput(userID, organizationID, spaceID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid list workbooks input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ListWorkbooks(ctx, input)
	if err != nil {
		handleWorkbookError(ctx, h.logger, c, "list workbooks", err)
		return
	}

	items := make([]api.WorkbookResponse, len(output.Workbooks))
	for i, wb := range output.Workbooks {
		items[i] = api.WorkbookResponse{
			WorkbookID:     wb.WorkbookID,
			SpaceID:        wb.SpaceID,
			OwnerID:        wb.OwnerID,
			OrganizationID: wb.OrganizationID,
			Title:          wb.Title,
			Description:    wb.Description,
			Visibility:     api.WorkbookResponseVisibility(wb.Visibility),
			Language:       wb.Language,
			CreatedAt:      wb.CreatedAt,
			UpdatedAt:      wb.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, api.ListWorkbooksResponse{
		Workbooks: items,
	})
}
