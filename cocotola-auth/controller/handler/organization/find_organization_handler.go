// Package organization provides HTTP handlers for organization lookup operations.
package organization

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// Finder finds an organization by name.
type Finder interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

// FindOrganizationHandler handles the GET /auth/organization endpoint.
type FindOrganizationHandler struct {
	orgFinder Finder
	logger    *slog.Logger
}

// NewFindOrganizationHandler returns a new FindOrganizationHandler.
func NewFindOrganizationHandler(orgFinder Finder) *FindOrganizationHandler {
	return &FindOrganizationHandler{
		orgFinder: orgFinder,
		logger:    slog.Default().With(slog.String(liblogging.LoggerNameKey, "FindOrganizationHandler")),
	}
}

// FindOrganization handles GET /auth/organization?name=<name>.
func (h *FindOrganizationHandler) FindOrganization(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.Query("name")
	if name == "" {
		h.logger.WarnContext(ctx, "missing name query parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "name query parameter is required"))
		return
	}

	org, err := h.orgFinder.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrOrganizationNotFound) {
			h.logger.WarnContext(ctx, "organization not found", slog.String("name", name))
			c.JSON(http.StatusNotFound, controller.NewErrorResponse("organization_not_found", "organization not found"))
			return
		}
		h.logger.ErrorContext(ctx, "find organization", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	// TODO(uuidv7-phase1-openapi): OpenAPI still encodes IDs as int32.
	orgID, err := handler.BridgeOrganizationIDToInt32(org.ID())
	if err != nil {
		h.logger.ErrorContext(ctx, "convert organization ID", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.FindOrganizationResponse{
		ID:   orgID,
		Name: org.Name(),
	})
}

// InitOrganizationRouter sets up the routes for organization operations.
func InitOrganizationRouter(
	findHandler *FindOrganizationHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	orgGroup := parentRouterGroup.Group("organization", middleware...)

	orgGroup.GET("", authMiddleware, findHandler.FindOrganization)
}
