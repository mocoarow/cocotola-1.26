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

// ShareWorkbookUsecase defines the use case method required by the ShareWorkbookHandler.
type ShareWorkbookUsecase interface {
	ShareWorkbook(ctx context.Context, input *referenceservice.ShareWorkbookInput) (*referenceservice.ShareWorkbookOutput, error)
}

// ShareWorkbookHandler handles the POST /workbook/:workbookId/share endpoint.
type ShareWorkbookHandler struct {
	usecase ShareWorkbookUsecase
	logger  *slog.Logger
}

// NewShareWorkbookHandler returns a new ShareWorkbookHandler.
func NewShareWorkbookHandler(usecase ShareWorkbookUsecase) *ShareWorkbookHandler {
	return &ShareWorkbookHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "ShareWorkbookHandler")),
	}
}

// ShareWorkbook handles POST /workbook/:workbookId/share.
func (h *ShareWorkbookHandler) ShareWorkbook(c *gin.Context) {
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

	input, err := referenceservice.NewShareWorkbookInput(userID, organizationID, workbookID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid share workbook input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	output, err := h.usecase.ShareWorkbook(ctx, input)
	if err != nil {
		handleSharingError(ctx, h.logger, c, "share workbook", err)
		return
	}

	c.JSON(http.StatusCreated, api.ShareWorkbookResponse{
		ReferenceID: output.ReferenceID,
		WorkbookID:  output.WorkbookID,
		AddedAt:     output.AddedAt,
	})
}
