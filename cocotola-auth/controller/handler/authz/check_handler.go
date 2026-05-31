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

// CheckHandler handles the /auth/authz/check endpoint (GET and POST).
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

// checkRequest is the JSON body for POST /auth/authz/check.
type checkRequest struct {
	Org      string `json:"org"`
	User     string `json:"user"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// Check handles POST /auth/authz/check with JSON body,
// or GET /auth/authz/check with query parameters (for backward compatibility).
func (h *CheckHandler) Check(c *gin.Context) {
	ctx := c.Request.Context()

	var orgStr, userStr, actionStr, resourceStr string

	if c.Request.Method == http.MethodPost {
		var req checkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.WarnContext(ctx, "invalid request body", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid request body"))
			return
		}
		orgStr = req.Org
		userStr = req.User
		actionStr = req.Action
		resourceStr = req.Resource
	} else {
		orgStr = c.Query("org")
		userStr = c.Query("user")
		actionStr = c.Query("action")
		resourceStr = c.Query("resource")
	}

	if orgStr == "" || userStr == "" || actionStr == "" || resourceStr == "" {
		h.logger.WarnContext(ctx, "missing required parameters")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "org, user, action, and resource are required"))
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
// Registers both GET (backward compatibility) and POST (preferred).
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
	authzGroup.POST("/check", handlers...)
}

// InitAuthzPolicyRouter sets up the routes for policy management operations (internal only).
func InitAuthzPolicyRouter(
	addPolicyHandler *AddPolicyHandler,
	parentRouterGroup gin.IRouter,
) {
	authzGroup := parentRouterGroup.Group("authz")
	authzGroup.POST("/policy", addPolicyHandler.AddPolicy)
}

// InitInternalAuthzRouter registers ALL internal (service-to-service) authz routes:
// the authorization check endpoint AND the policy-management endpoint. Internal
// callers such as cocotola-question need both (creating a workbook performs an
// authz check and then grants per-resource policies), so they are registered
// together here. Both the standalone-deployment wiring (cocotola-auth/main.go)
// and the monolith wiring (cocotola-auth/initialize) MUST call this single
// function so the two paths cannot drift — a past bug registered /check
// internally but omitted /policy, making policy grants 404 in production.
func InitInternalAuthzRouter(
	checkHandler *CheckHandler,
	addPolicyHandler *AddPolicyHandler,
	parentRouterGroup gin.IRouter,
) {
	InitAuthzRouter(checkHandler, parentRouterGroup)
	InitAuthzPolicyRouter(addPolicyHandler, parentRouterGroup)
}
