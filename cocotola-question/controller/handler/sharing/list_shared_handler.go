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

// ListSharedUsecase defines the use case method required by the ListSharedHandler.
type ListSharedUsecase interface {
	ListShared(ctx context.Context, input *referenceservice.ListSharedInput) (*referenceservice.ListSharedOutput, error)
}

// ListSharedHandler handles the GET /workbook/shared endpoint.
type ListSharedHandler struct {
	usecase ListSharedUsecase
	logger  *slog.Logger
}

// NewListSharedHandler returns a new ListSharedHandler.
func NewListSharedHandler(usecase ListSharedUsecase) *ListSharedHandler {
	return &ListSharedHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ListSharedHandler")),
	}
}

// ListShared handles GET /workbook/shared.
func (h *ListSharedHandler) ListShared(c *gin.Context) {
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

	input, err := referenceservice.NewListSharedInput(userID, organizationID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid list shared input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ListShared(ctx, input)
	if err != nil {
		handleSharingError(ctx, h.logger, c, "list shared", err)
		return
	}

	items := make([]api.SharedItemResponse, len(output.References))
	for i, ref := range output.References {
		items[i] = api.SharedItemResponse{
			ReferenceID: ref.ReferenceID,
			WorkbookID:  ref.WorkbookID,
			AddedAt:     ref.AddedAt,
		}
	}

	c.JSON(http.StatusOK, api.ListSharedResponse{
		References: items,
	})
}
