package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"

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
	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
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

	c.JSON(http.StatusOK, api.GetMeResponse{
		UserID:           userID.UUID(),
		LoginID:          loginID,
		OrganizationName: organizationName,
	})
}
