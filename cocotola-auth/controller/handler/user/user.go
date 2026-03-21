package user

import (
	"github.com/gin-gonic/gin"
)

// InitUserRouter sets up the routes for user operations under the given parent router group.
func InitUserRouter(
	createUserHandler *CreateUserHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	userGroup := parentRouterGroup.Group("user", middleware...)

	userGroup.POST("", authMiddleware, createUserHandler.CreateUser)
}
