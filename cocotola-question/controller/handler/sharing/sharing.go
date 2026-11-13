// Package sharing provides HTTP handlers for workbook sharing operations.
package sharing

import (
	"github.com/gin-gonic/gin"
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
