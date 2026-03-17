package auth

import (
	"github.com/gin-gonic/gin"
)

// InitAuthRouter sets up the routes for auth operations under the given parent router group.
func InitAuthRouter(
	authenticateHandler *PasswordAuthenticateHandler,
	refreshHandler *RefreshHandler,
	revokeHandler *RevokeHandler,
	getMeHandler *GetMeHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	authGroup := parentRouterGroup.Group("auth", middleware...)

	authGroup.POST("/authenticate", authenticateHandler.Authenticate)
	authGroup.POST("/logout", revokeHandler.Logout)
	authGroup.POST("/refresh", refreshHandler.Refresh)
	authGroup.POST("/revoke", authMiddleware, revokeHandler.Revoke)
	authGroup.GET("/me", authMiddleware, getMeHandler.GetMe)
}
