package auth

import (
	"github.com/gin-gonic/gin"
)

// InitAuthRouter sets up the routes for auth operations under the given parent router group.
func InitAuthRouter(
	authenticateHandler *PasswordAuthenticateHandler,
	guestAuthenticateHandler *GuestAuthenticateHandler,
	refreshHandler *RefreshHandler,
	revokeHandler *RevokeHandler,
	getMeHandler *GetMeHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	authGroup := parentRouterGroup.Group("auth", middleware...)

	authGroup.POST("/authenticate", authenticateHandler.Authenticate)
	authGroup.POST("/guest/authenticate", guestAuthenticateHandler.Authenticate)
	authGroup.POST("/logout", revokeHandler.Logout)
	authGroup.POST("/refresh", refreshHandler.Refresh)
	authGroup.POST("/revoke", authMiddleware, revokeHandler.Revoke)
	authGroup.GET("/me", authMiddleware, getMeHandler.GetMe)
}

// InitInternalAuthRouter sets up the routes for internal (service-to-service) auth operations.
// These routes are protected by API key middleware applied at the parent group level.
func InitInternalAuthRouter(
	supabaseExchangeHandler *SupabaseExchangeHandler,
	parentRouterGroup gin.IRouter,
) {
	authGroup := parentRouterGroup.Group("auth")

	authGroup.POST("/supabase/exchange", supabaseExchangeHandler.Exchange)
}
