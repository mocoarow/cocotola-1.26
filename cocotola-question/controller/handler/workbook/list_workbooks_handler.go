package workbook

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
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

	spaceIDStr := c.Query("spaceId")
	if spaceIDStr == "" {
		h.logger.WarnContext(ctx, "missing spaceId query parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "spaceId query parameter is required"))
		return
	}

	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid spaceId query parameter", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "spaceId must be a valid integer"))
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
		sid, err := handler.SafeIntToInt32(wb.SpaceID)
		if err != nil {
			h.logger.ErrorContext(ctx, "convert space ID", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			return
		}
		oid, err := handler.SafeIntToInt32(wb.OwnerID)
		if err != nil {
			h.logger.ErrorContext(ctx, "convert owner ID", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			return
		}
		orgID, err := handler.SafeIntToInt32(wb.OrganizationID)
		if err != nil {
			h.logger.ErrorContext(ctx, "convert organization ID", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			return
		}
		items[i] = api.WorkbookResponse{
			WorkbookID:     wb.WorkbookID,
			SpaceID:        sid,
			OwnerID:        oid,
			OrganizationID: orgID,
			Title:          wb.Title,
			Description:    wb.Description,
			Visibility:     api.WorkbookResponseVisibility(wb.Visibility),
			CreatedAt:      wb.CreatedAt,
			UpdatedAt:      wb.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, api.ListWorkbooksResponse{
		Workbooks: items,
	})
}
