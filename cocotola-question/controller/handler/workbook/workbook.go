// Package workbook provides HTTP handlers for workbook management operations.
package workbook

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// InitWorkbookRouter sets up the routes for workbook operations under the given parent router group.
func InitWorkbookRouter(
	createHandler *CreateWorkbookHandler,
	getHandler *GetWorkbookHandler,
	listHandler *ListWorkbooksHandler,
	updateHandler *UpdateWorkbookHandler,
	deleteHandler *DeleteWorkbookHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	workbookGroup := parentRouterGroup.Group("workbook", middleware...)

	workbookGroup.POST("", authMiddleware, createHandler.CreateWorkbook)
	workbookGroup.GET("", authMiddleware, listHandler.ListWorkbooks)
	workbookGroup.GET("/:workbookId", authMiddleware, getHandler.GetWorkbook)
	workbookGroup.PUT("/:workbookId", authMiddleware, updateHandler.UpdateWorkbook)
	workbookGroup.DELETE("/:workbookId", authMiddleware, deleteHandler.DeleteWorkbook)
}

func handleWorkbookError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
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
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
