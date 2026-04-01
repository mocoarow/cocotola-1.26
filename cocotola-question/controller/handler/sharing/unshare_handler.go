package sharing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// UnshareUsecase defines the use case method required by the UnshareHandler.
type UnshareUsecase interface {
	Unshare(ctx context.Context, input *referenceservice.UnshareInput) error
}

// UnshareHandler handles the DELETE /workbook/shared/:refId endpoint.
type UnshareHandler struct {
	usecase UnshareUsecase
	logger  *slog.Logger
}

// NewUnshareHandler returns a new UnshareHandler.
func NewUnshareHandler(usecase UnshareUsecase) *UnshareHandler {
	return &UnshareHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "UnshareHandler")),
	}
}

// Unshare handles DELETE /workbook/shared/:refId.
func (h *UnshareHandler) Unshare(c *gin.Context) {
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

	refID := c.Param("refId")
	if refID == "" {
		h.logger.WarnContext(ctx, "missing reference ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "reference ID is required"))
		return
	}

	input, err := referenceservice.NewUnshareInput(userID, organizationID, refID)
	if err != nil {
		h.logger.ErrorContext(ctx, "invalid unshare input", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err := h.usecase.Unshare(ctx, input); err != nil {
		handleSharingError(ctx, h.logger, c, "unshare", err)
		return
	}

	c.Status(http.StatusNoContent)
}
