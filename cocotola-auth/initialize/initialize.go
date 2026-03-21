// Package initialize provides a reusable initialization function for the cocotola-auth module.
package initialize

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/config"
	authhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/auth"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

const eventBusBufferSize = 100

// Initialize sets up the cocotola-auth module: gateway, usecase, and controller layers.
// It registers all auth-related routes under the given parent router group and returns
// a RunProcessFunc for the event bus that the caller should pass to libprocess.Run.
func Initialize(_ context.Context, parent gin.IRouter, db *gorm.DB, authConfig config.AuthConfig) (libprocess.RunProcessFunc, error) {
	// gateway layer
	jwtManager := gateway.NewJWTManager(
		[]byte(authConfig.SigningKey),
		jwt.SigningMethodHS256,
		time.Duration(authConfig.AccessTokenTTLMin)*time.Minute,
	)
	bcryptHasher := gateway.NewBcryptHasher()
	rbacRepo, err := gateway.NewRBACRepository(db)
	if err != nil {
		return nil, fmt.Errorf("new RBAC repository: %w", err)
	}
	userAuthenticator := gateway.NewUserAuthenticator(db, bcryptHasher, rbacRepo)
	sessionTokenRepo := gateway.NewSessionTokenRepository(db)
	sessionTokenWhitelistRepo := gateway.NewSessionTokenWhitelistRepository(db)
	refreshTokenRepo := gateway.NewRefreshTokenRepository(db)
	refreshTokenWhitelistRepo := gateway.NewRefreshTokenWhitelistRepository(db)
	accessTokenRepo := gateway.NewAccessTokenRepository(db)
	accessTokenWhitelistRepo := gateway.NewAccessTokenWhitelistRepository(db)
	tokenCache := gateway.NewTokenCache()

	// event bus
	eventBusLogger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-event-bus"))
	eventBus := gateway.NewEventBus(eventBusBufferSize, eventBusLogger)

	activeUserListRepo := gateway.NewActiveUserListRepository(db)
	orgRepo := gateway.NewOrganizationRepository(db)
	eventHandlerLogger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-event-handler"))
	activeUserListHandler := eventusecase.NewActiveUserListHandler(activeUserListRepo, orgRepo, eventHandlerLogger)
	eventBus.Subscribe(domain.EventTypeAppUserCreated, activeUserListHandler.Handle)

	// usecase layer
	usecaseConfig := authusecase.UsecaseConfig{
		SessionTokenTTLMin: authConfig.SessionTokenTTLMin,
		SessionMaxTTLMin:   authConfig.SessionMaxTTLMin,
		AccessTokenTTLMin:  authConfig.AccessTokenTTLMin,
		RefreshTokenTTLMin: authConfig.RefreshTokenTTLMin,
		TokenWhitelistSize: authConfig.TokenWhitelistSize,
		ClockFunc:          nil,
	}
	authUsecase := authusecase.NewUsecase(
		userAuthenticator,
		sessionTokenRepo,
		sessionTokenWhitelistRepo,
		refreshTokenRepo,
		refreshTokenWhitelistRepo,
		accessTokenRepo,
		accessTokenWhitelistRepo,
		jwtManager,
		tokenCache,
		usecaseConfig,
	)

	// controller layer
	api := parent.Group("api")
	v1 := api.Group("v1")

	authMiddleware := middleware.NewAuthMiddleware(authUsecase, authConfig.Cookie, authConfig.SessionTokenTTLMin)
	authenticateHandler := authhandler.NewPasswordAuthenticateHandler(authUsecase, authConfig.Cookie, authConfig.SessionTokenTTLMin)
	refreshHandler := authhandler.NewRefreshHandler(authUsecase)
	revokeHandler := authhandler.NewRevokeHandler(authUsecase, authConfig.Cookie)
	getMeHandler := authhandler.NewGetMeHandler()
	authhandler.InitAuthRouter(authenticateHandler, refreshHandler, revokeHandler, getMeHandler, v1, authMiddleware)

	return eventBus.Start, nil
}
