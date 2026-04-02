// Package sharing provides HTTP handlers for workbook sharing operations.
package sharing

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// InitSharingRouter sets up the routes for sharing operations under the given parent router group.
func InitSharingRouter(
	shareHandler *ShareWorkbookHandler,
	listSharedHandler *ListSharedHandler,
	unshareHandler *UnshareHandler,
	listPublicHandler *ListPublicHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	sharingGroup := parentRouterGroup.Group("workbook")
	sharingGroup.Use(authMiddleware)
	sharingGroup.Use(middleware...)

	sharingGroup.POST("/:workbookId/share", shareHandler.ShareWorkbook)
	sharingGroup.GET("/shared", listSharedHandler.ListShared)
	sharingGroup.DELETE("/shared/:refId", unshareHandler.Unshare)
	sharingGroup.GET("/public", listPublicHandler.ListPublic)
}

func handleSharingError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		logger.WarnContext(ctx, "forbidden", slog.Any("error", err))
		c.JSON(http.StatusForbidden, controller.NewErrorResponse("forbidden", http.StatusText(http.StatusForbidden)))
		return
	}
	if errors.Is(err, domain.ErrWorkbookNotFound) {
		logger.WarnContext(ctx, "workbook not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("workbook_not_found", "workbook not found"))
		return
	}
	if errors.Is(err, domain.ErrReferenceNotFound) {
		logger.WarnContext(ctx, "reference not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("reference_not_found", "workbook reference not found"))
		return
	}
	if errors.Is(err, domain.ErrDuplicateReference) {
		logger.WarnContext(ctx, "duplicate reference", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("duplicate_reference", "workbook reference already exists"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
