package auth

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// GetMeHandler handles the GET /auth/me endpoint.
type GetMeHandler struct {
	logger *slog.Logger
}

// NewGetMeHandler returns a new GetMeHandler.
func NewGetMeHandler() *GetMeHandler {
	return &GetMeHandler{
		logger: slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetMeHandler")),
	}
}

// GetMe handles GET /auth/me and returns the authenticated user's ID and login ID.
func (h *GetMeHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt(controller.ContextFieldUserID{})
	if userID <= 0 {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	loginID := c.GetString(controller.ContextFieldLoginID{})
	if loginID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing login ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationName := c.GetString(controller.ContextFieldOrganizationName{})

	userIDInt32, err := safeIntToInt32(userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "convert user ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.GetMeResponse{
		UserID:           userIDInt32,
		LoginID:          loginID,
		OrganizationName: organizationName,
	})
}

func safeIntToInt32(v int) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d overflows int32", v)
	}
	return int32(v), nil
}
