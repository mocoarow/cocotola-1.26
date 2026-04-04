// Package main is the entry point for the cocotola-auth microservice.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/config"
	authhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/auth"
	authzhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/authz"
	healthhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/health"
	orghandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/organization"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"
)

const eventBusBufferSize = 100

func main() {
	exitCode, err := run()
	if err != nil {
		slog.Error("run", slog.Any("error", err))
	}
	os.Exit(exitCode)
}

func run() (int, error) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		return 1, fmt.Errorf("load config: %w", err)
	}
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-main"))

	// init log
	shutdownlog, err := libgateway.InitLog(ctx, cfg.Log, domain.AppName)
	if err != nil {
		return 0, fmt.Errorf("init log: %w", err)
	}
	defer shutdownlog()

	// init tracer
	shutdownTrace, err := libgateway.InitTracerProvider(ctx, cfg.Trace, domain.AppName)
	if err != nil {
		return 0, fmt.Errorf("init trace: %w", err)
	}
	defer shutdownTrace()

	// init db
	dbConn, shutdownDB, err := libgateway.InitDB(ctx, cfg.DB, cfg.Log, domain.AppName)
	if err != nil {
		return 1, fmt.Errorf("init db: %w", err)
	}
	defer shutdownDB()

	router, err := libhandler.InitRootRouterGroup(ctx, cfg.Server, domain.AppName)
	if err != nil {
		return 1, fmt.Errorf("init router: %w", err)
	}

	// gateway layer
	jwtManager := gateway.NewJWTManager(
		[]byte(cfg.Auth.SigningKey),
		jwt.SigningMethodHS256,
		time.Duration(cfg.Auth.AccessTokenTTLMin)*time.Minute,
	)
	bcryptHasher := gateway.NewBcryptHasher()
	rbacRepo, err := gateway.NewRBACRepository(dbConn.DB)
	if err != nil {
		return 1, fmt.Errorf("new rbac repository: %w", err)
	}
	userAuthenticator := gateway.NewUserAuthenticator(dbConn.DB, bcryptHasher, rbacRepo)
	guestAuthenticator := gateway.NewGuestAuthenticator(dbConn.DB)
	sessionTokenRepo := gateway.NewSessionTokenRepository(dbConn.DB)
	sessionTokenWhitelistRepo := gateway.NewSessionTokenWhitelistRepository(dbConn.DB)
	refreshTokenRepo := gateway.NewRefreshTokenRepository(dbConn.DB)
	refreshTokenWhitelistRepo := gateway.NewRefreshTokenWhitelistRepository(dbConn.DB)
	accessTokenRepo := gateway.NewAccessTokenRepository(dbConn.DB)
	accessTokenWhitelistRepo := gateway.NewAccessTokenWhitelistRepository(dbConn.DB)
	tokenCache := gateway.NewTokenCache()

	// event bus
	eventBusLogger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-event-bus"))
	eventBus := gateway.NewEventBus(eventBusBufferSize, eventBusLogger)

	activeUserListRepo := gateway.NewActiveUserListRepository(dbConn.DB)
	orgRepo := gateway.NewOrganizationRepository(dbConn.DB)
	eventHandlerLogger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-event-handler"))
	activeUserListHandler := eventusecase.NewActiveUserListHandler(activeUserListRepo, orgRepo, eventHandlerLogger)
	eventBus.Subscribe(domain.EventTypeAppUserCreated, activeUserListHandler.Handle)

	// usecase layer
	usecaseConfig := authusecase.UsecaseConfig{
		SessionTokenTTLMin: cfg.Auth.SessionTokenTTLMin,
		SessionMaxTTLMin:   cfg.Auth.SessionMaxTTLMin,
		AccessTokenTTLMin:  cfg.Auth.AccessTokenTTLMin,
		RefreshTokenTTLMin: cfg.Auth.RefreshTokenTTLMin,
		TokenWhitelistSize: cfg.Auth.TokenWhitelistSize,
		ClockFunc:          nil,
	}
	authUsecase := authusecase.NewUsecase(
		userAuthenticator,
		guestAuthenticator,
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

	// api
	api := router.Group("api")

	// v1
	v1 := api.Group("v1")

	// health check
	{
		healthRepo := gateway.NewHealthRepository(dbConn.DB)
		checkHandler := healthhandler.NewCheckHandler(healthRepo)
		healthhandler.InitRouter(checkHandler, v1)
	}

	authMiddleware := middleware.NewAuthMiddleware(authUsecase, cfg.Auth.Cookie, cfg.Auth.SessionTokenTTLMin)
	{
		authenticateHandler := authhandler.NewPasswordAuthenticateHandler(authUsecase, cfg.Auth.Cookie, cfg.Auth.SessionTokenTTLMin)
		guestAuthenticateHandler := authhandler.NewGuestAuthenticateHandler(authUsecase)
		refreshHandler := authhandler.NewRefreshHandler(authUsecase)
		revokeHandler := authhandler.NewRevokeHandler(authUsecase, cfg.Auth.Cookie)
		getMeHandler := authhandler.NewGetMeHandler()
		authhandler.InitAuthRouter(authenticateHandler, guestAuthenticateHandler, refreshHandler, revokeHandler, getMeHandler, v1, authMiddleware)
	}

	authV1 := v1.Group("auth")

	// organization lookup
	{
		findOrgHandler := orghandler.NewFindOrganizationHandler(orgRepo)
		orghandler.InitOrganizationRouter(findOrgHandler, authV1, authMiddleware)
	}

	// authz check
	{
		authzChecker := gateway.NewCasbinAuthorizationChecker(rbacRepo)
		authzCheckHandler := authzhandler.NewCheckHandler(authzChecker)
		authzhandler.InitAuthzRouter(authzCheckHandler, authV1, authMiddleware)
	}

	// run
	readHeaderTimeout := time.Duration(cfg.Server.ReadHeaderTimeoutSec) * time.Second
	shutdownTime := time.Duration(cfg.Server.Shutdown.ShutdownTimeSec) * time.Second
	result := libprocess.Run(ctx,
		libcontroller.WithWebServerProcess(router, cfg.Server.HTTPPort, readHeaderTimeout, shutdownTime),
		libcontroller.WithMetricsServerProcess(cfg.Server.MetricsPort, readHeaderTimeout, shutdownTime),
		libgateway.WithSignalWatchProcess(),
		eventBus.Start,
	)

	gracefulShutdownTime2 := time.Duration(cfg.Server.Shutdown.GracePeriodSec) * time.Second
	time.Sleep(gracefulShutdownTime2)
	logger.InfoContext(ctx, "exited")
	return result, nil
}
