package sharing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// ListPublicUsecase defines the use case method required by the ListPublicHandler.
type ListPublicUsecase interface {
	ListPublic(ctx context.Context, input *referenceservice.ListPublicInput) (*referenceservice.ListPublicOutput, error)
}

// ListPublicHandler handles the GET /workbook/public endpoint.
type ListPublicHandler struct {
	usecase ListPublicUsecase
	logger  *slog.Logger
}

// NewListPublicHandler returns a new ListPublicHandler.
func NewListPublicHandler(usecase ListPublicUsecase) *ListPublicHandler {
	return &ListPublicHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListPublicHandler")),
	}
}

// ListPublic handles GET /workbook/public.
func (h *ListPublicHandler) ListPublic(c *gin.Context) {
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

	input, err := referenceservice.NewListPublicInput(userID, organizationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid list public input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ListPublic(ctx, input)
	if err != nil {
		handleSharingError(ctx, h.logger, c, "list public", err)
		return
	}

	items := make([]api.PublicWorkbookResponse, len(output.Workbooks))
	for i, wb := range output.Workbooks {
		items[i] = api.PublicWorkbookResponse{
			WorkbookID:  wb.WorkbookID,
			OwnerID:     wb.OwnerID,
			Title:       wb.Title,
			Description: wb.Description,
			CreatedAt:   wb.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, api.ListPublicResponse{
		Workbooks: items,
	})
}
