package health

import (
	"github.com/gin-gonic/gin"
)

// InitRouter sets up the routes for health operations under the given parent router group.
func InitRouter(
	checkHandler *CheckHandler,
	parentRouterGroup gin.IRouter,
	middleware ...gin.HandlerFunc,
) {
	authGroup := parentRouterGroup.Group("auth", middleware...)

	authGroup.GET("/health", checkHandler.HealthCheck)
}
