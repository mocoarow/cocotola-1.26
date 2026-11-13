// Package workbook provides HTTP handlers for workbook management operations.
package workbook

import (
	"github.com/gin-gonic/gin"
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
// parent router (protected by X-Service-Api-Key).
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
