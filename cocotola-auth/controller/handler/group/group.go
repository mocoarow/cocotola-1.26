package group

import (
	"github.com/gin-gonic/gin"
)

// InitGroupRouter sets up the routes for group operations under the given parent router group.
func InitGroupRouter(
	createGroupHandler *CreateGroupHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	groupGroup := parentRouterGroup.Group("group", middleware...)

	groupGroup.POST("", authMiddleware, createGroupHandler.CreateGroup)
}
