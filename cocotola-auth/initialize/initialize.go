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
	authzhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/authz"
	grouphandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/group"
	healthhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/health"
	orghandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/organization"
	spacehandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/space"
	userhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"
	groupusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/group"
	spaceusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/space"
	userusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/user"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

const eventBusBufferSize = 100

// AuthorizationChecker checks if an action is allowed by RBAC policy.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}

// OrganizationFinder finds organizations by name.
type OrganizationFinder interface {
	FindByName(ctx context.Context, name string) (*domain.Organization, error)
}

// InitResult holds the results of auth module initialization for use by other modules.
type InitResult struct {
	// EventBusStart is the RunProcessFunc for the event bus.
	EventBusStart libprocess.RunProcessFunc
	// AuthMiddleware is the Gin middleware for authenticating requests.
	AuthMiddleware gin.HandlerFunc
	// V1RouterGroup is the /api/v1 router group for registering additional routes.
	V1RouterGroup gin.IRouter
	// AuthzChecker is the RBAC authorization checker for use by other modules.
	AuthzChecker AuthorizationChecker
	// OrgFinder finds organizations by name for use by other modules.
	OrgFinder OrganizationFinder
}

// Initialize sets up the cocotola-auth module: gateway, usecase, and controller layers.
// It registers all auth-related routes under the given parent router group and returns
// an InitResult containing shared resources for use by other modules.
func Initialize(_ context.Context, parent gin.IRouter, db *gorm.DB, authConfig config.AuthConfig, internalConfig config.InternalConfig) (*InitResult, error) {
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
	guestAuthenticator := gateway.NewGuestAuthenticator(db)
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
	activeGroupListRepo := gateway.NewActiveGroupListRepository(db)
	orgRepo := gateway.NewOrganizationRepository(db)
	groupRepo := gateway.NewGroupRepository(db)
	eventHandlerLogger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-event-handler"))
	activeUserListHandler := eventusecase.NewActiveUserListHandler(activeUserListRepo, orgRepo, eventHandlerLogger)
	eventBus.Subscribe(domain.EventTypeAppUserCreated, activeUserListHandler.Handle)
	activeGroupListHandler := eventusecase.NewActiveGroupListHandler(activeGroupListRepo, orgRepo, eventHandlerLogger)
	eventBus.Subscribe(domain.EventTypeGroupCreated, activeGroupListHandler.Handle)
	spaceRepo := gateway.NewSpaceRepository(db)
	privateSpaceHandler := eventusecase.NewPrivateSpaceHandler(spaceRepo, rbacRepo, eventHandlerLogger)
	eventBus.Subscribe(domain.EventTypeAppUserCreated, privateSpaceHandler.Handle)

	// usecase layer
	usecaseConfig := authusecase.UsecaseConfig{
		SessionTokenTTLMin: authConfig.SessionTokenTTLMin,
		SessionMaxTTLMin:   authConfig.SessionMaxTTLMin,
		AccessTokenTTLMin:  authConfig.AccessTokenTTLMin,
		RefreshTokenTTLMin: authConfig.RefreshTokenTTLMin,
		TokenWhitelistSize: authConfig.TokenWhitelistSize,
		ClockFunc:          nil,
	}
	supabaseVerifier := gateway.NewSupabaseVerifier(authConfig.Supabase.JWTSecret)
	appUserRepo := gateway.NewAppUserRepository(db)
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
		supabaseVerifier,
		appUserRepo,
		appUserRepo,
		orgRepo,
	)

	// controller layer
	api := parent.Group("api")
	v1 := api.Group("v1")

	// health check
	healthRepo := gateway.NewHealthRepository(db)
	checkHandler := healthhandler.NewCheckHandler(healthRepo)
	healthhandler.InitRouter(checkHandler, v1)

	authMiddleware := middleware.NewAuthMiddleware(authUsecase, authConfig.Cookie, authConfig.SessionTokenTTLMin)
	authenticateHandler := authhandler.NewPasswordAuthenticateHandler(authUsecase, authConfig.Cookie, authConfig.SessionTokenTTLMin)
	guestAuthenticateHandler := authhandler.NewGuestAuthenticateHandler(authUsecase)
	refreshHandler := authhandler.NewRefreshHandler(authUsecase)
	revokeHandler := authhandler.NewRevokeHandler(authUsecase, authConfig.Cookie)
	getMeHandler := authhandler.NewGetMeHandler()
	authhandler.InitAuthRouter(authenticateHandler, guestAuthenticateHandler, refreshHandler, revokeHandler, getMeHandler, v1, authMiddleware)

	// internal (service-to-service) routes protected by API key
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(internalConfig.APIKey)
	internalV1 := api.Group("v1/internal", apiKeyMiddleware)
	supabaseExchangeHandler := authhandler.NewSupabaseExchangeHandler(authUsecase)
	authhandler.InitInternalAuthRouter(supabaseExchangeHandler, internalV1)

	// group usecase + controller
	authzChecker := gateway.NewCasbinAuthorizationChecker(rbacRepo)
	authV1 := v1.Group("auth")
	groupCommand := groupusecase.NewCommand(groupRepo, orgRepo, eventBus, authzChecker)
	createGroupHandler := grouphandler.NewCreateGroupHandler(groupCommand)
	grouphandler.InitGroupRouter(createGroupHandler, authV1, authMiddleware)

	// space usecase + controller
	spaceCommand := spaceusecase.NewCommand(spaceRepo, spaceRepo, orgRepo, eventBus, authzChecker)
	createSpaceHandler := spacehandler.NewCreateSpaceHandler(spaceCommand)
	listSpacesHandler := spacehandler.NewListSpacesHandler(spaceCommand)
	spacehandler.InitSpaceRouter(createSpaceHandler, listSpacesHandler, authV1, authMiddleware)

	// user usecase + controller
	userCommand := userusecase.NewCommand(appUserRepo, orgRepo, eventBus, appUserRepo, appUserRepo, bcryptHasher, authzChecker)
	createUserHandler := userhandler.NewCreateUserHandler(userCommand)
	userhandler.InitUserRouter(createUserHandler, authV1, authMiddleware)

	// organization lookup + controller
	findOrgHandler := orghandler.NewFindOrganizationHandler(orgRepo)
	orghandler.InitOrganizationRouter(findOrgHandler, authV1, authMiddleware)

	// authz check + controller
	authzCheckHandler := authzhandler.NewCheckHandler(authzChecker)
	authzhandler.InitAuthzRouter(authzCheckHandler, authV1, authMiddleware)

	return &InitResult{
		EventBusStart:  eventBus.Start,
		AuthMiddleware: authMiddleware,
		V1RouterGroup:  v1,
		AuthzChecker:   authzChecker,
		OrgFinder:      orgRepo,
	}, nil
}
