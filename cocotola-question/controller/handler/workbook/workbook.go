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
	workbookGroup := parentRouterGroup.Group("workbook")
	workbookGroup.Use(authMiddleware)
	workbookGroup.Use(middleware...)

	workbookGroup.POST("", createHandler.CreateWorkbook)
	workbookGroup.GET("", listHandler.ListWorkbooks)
	workbookGroup.GET("/:workbookId", getHandler.GetWorkbook)
	workbookGroup.PUT("/:workbookId", updateHandler.UpdateWorkbook)
	workbookGroup.DELETE("/:workbookId", deleteHandler.DeleteWorkbook)
}

// InitInternalWorkbookRouter mounts workbook CRUD endpoints on the internal
// parent router (which is expected to be protected by X-Service-Api-Key).
// The handler funcs are shared with the public router; the only difference is
// that operatorID/organizationID are populated from the API-key middleware.
//
// The list endpoint is included so that batch jobs (e.g. cocotola-init seeding)
// can perform idempotency checks before creating workbooks.
func InitInternalWorkbookRouter(
	createHandler *CreateWorkbookHandler,
	listHandler *ListWorkbooksHandler,
	updateHandler *UpdateWorkbookHandler,
	deleteHandler *DeleteWorkbookHandler,
	parentRouterGroup gin.IRouter,
) {
	workbookGroup := parentRouterGroup.Group("workbook")
	workbookGroup.POST("", createHandler.CreateWorkbook)
	workbookGroup.GET("", listHandler.ListWorkbooks)
	workbookGroup.PUT("/:workbookId", updateHandler.UpdateWorkbook)
	workbookGroup.DELETE("/:workbookId", deleteHandler.DeleteWorkbook)
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
	if errors.Is(err, domain.ErrOwnedWorkbookListNotFound) {
		logger.WarnContext(ctx, "owned workbook list not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("owned_workbook_list_not_found", "owned workbook list not found"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
