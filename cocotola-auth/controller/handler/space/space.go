// Package space provides HTTP handlers for space management operations.
package space

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// InitSpaceRouter sets up the routes for space operations under the given parent router group.
func InitSpaceRouter(
	createSpaceHandler *CreateSpaceHandler,
	listSpacesHandler *ListSpacesHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	spaceGroup := parentRouterGroup.Group("space", middleware...)

	spaceGroup.POST("", authMiddleware, createSpaceHandler.CreateSpace)
	spaceGroup.GET("", authMiddleware, listSpacesHandler.ListSpaces)
}

// InitInternalSpaceRouter wires the internal (service-to-service) space routes
// under the given parent router. The parent is expected to be already protected
// by the X-Service-Api-Key middleware.
func InitInternalSpaceRouter(
	findSpaceHandler *FindSpaceHandler,
	parentRouterGroup gin.IRouter,
) {
	spaceGroup := parentRouterGroup.Group("space")
	spaceGroup.GET("/:spaceId", findSpaceHandler.FindSpace)
}

func handleSpaceError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		logger.WarnContext(ctx, "forbidden", slog.Any("error", err))
		c.JSON(http.StatusForbidden, controller.NewErrorResponse("forbidden", http.StatusText(http.StatusForbidden)))
		return
	}
	if errors.Is(err, domain.ErrOrganizationNotFound) {
		logger.WarnContext(ctx, "organization not found", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("organization_not_found", "organization not found"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
