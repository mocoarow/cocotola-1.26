// Package authz provides HTTP handlers for authorization check operations.
package authz

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// AuthorizationChecker checks if an action is allowed by RBAC policy.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID domain.OrganizationID, operatorID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}

// CheckHandler handles the GET /auth/authz/check endpoint.
type CheckHandler struct {
	authzChecker AuthorizationChecker
	logger       *slog.Logger
}

// NewCheckHandler returns a new CheckHandler.
func NewCheckHandler(authzChecker AuthorizationChecker) *CheckHandler {
	return &CheckHandler{
		authzChecker: authzChecker,
		logger:       slog.Default().With(slog.String(liblogging.LoggerNameKey, "AuthzCheckHandler")),
	}
}

// Check handles GET /auth/authz/check?org=<orgID>&user=<userID>&action=<action>&resource=<resource>.
func (h *CheckHandler) Check(c *gin.Context) {
	ctx := c.Request.Context()

	orgStr := c.Query("org")
	userStr := c.Query("user")
	actionStr := c.Query("action")
	resourceStr := c.Query("resource")

	if orgStr == "" || userStr == "" || actionStr == "" || resourceStr == "" {
		h.logger.WarnContext(ctx, "missing required query parameters")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "org, user, action, and resource query parameters are required"))
		return
	}

	orgID, err := domain.ParseOrganizationID(orgStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid org parameter", slog.String("org", orgStr))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "org must be a UUID"))
		return
	}

	userID, err := domain.ParseAppUserID(userStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user parameter", slog.String("user", userStr))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "user must be a UUID"))
		return
	}

	action, err := domainrbac.NewAction(actionStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid action parameter", slog.String("action", actionStr))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid action"))
		return
	}

	resource, err := domainrbac.NewResource(resourceStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid resource parameter", slog.String("resource", resourceStr))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid resource"))
		return
	}

	allowed, err := h.authzChecker.IsAllowed(ctx, orgID, userID, action, resource)
	if err != nil {
		h.logger.ErrorContext(ctx, "authorization check", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.AuthzCheckResponse{
		Allowed: allowed,
	})
}

// InitAuthzRouter sets up the routes for authorization check operations.
// When middleware is provided, it is used as per-route auth middleware.
func InitAuthzRouter(
	checkHandler *CheckHandler,
	parentRouterGroup gin.IRouter,
	middleware ...gin.HandlerFunc,
) {
	authzGroup := parentRouterGroup.Group("authz")

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, checkHandler.Check)
	authzGroup.GET("/check", handlers...)
}
