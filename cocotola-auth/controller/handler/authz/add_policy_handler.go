package authz

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// UserPolicyAdder adds per-user RBAC policies.
type UserPolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
}

// AddPolicyHandler handles POST /auth/authz/policy.
type AddPolicyHandler struct {
	policyAdder UserPolicyAdder
	logger      *slog.Logger
}

// NewAddPolicyHandler returns a new AddPolicyHandler.
func NewAddPolicyHandler(policyAdder UserPolicyAdder) *AddPolicyHandler {
	return &AddPolicyHandler{
		policyAdder: policyAdder,
		logger:      slog.Default().With(slog.String(liblogging.LoggerNameKey, "AddPolicyHandler")),
	}
}

// addPolicyRequest is the JSON body for POST /auth/authz/policy.
type addPolicyRequest struct {
	Org      string `json:"org"`
	User     string `json:"user"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Effect   string `json:"effect"`
}

type addPolicyParams struct {
	orgID    domain.OrganizationID
	userID   domain.AppUserID
	action   domainrbac.Action
	resource domainrbac.Resource
	effect   domainrbac.Effect
}

func (h *AddPolicyHandler) parseRequest(ctx context.Context, c *gin.Context) (*addPolicyParams, bool) {
	var req addPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid request body"))
		return nil, false
	}

	if req.Org == "" || req.User == "" || req.Action == "" || req.Resource == "" || req.Effect == "" {
		h.logger.WarnContext(ctx, "missing required parameters")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "org, user, action, resource, and effect are required"))
		return nil, false
	}

	orgID, err := domain.ParseOrganizationID(req.Org)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid org parameter", slog.String("org", req.Org))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "org must be a UUID"))
		return nil, false
	}

	userID, err := domain.ParseAppUserID(req.User)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user parameter", slog.String("user", req.User))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "user must be a UUID"))
		return nil, false
	}

	action, err := domainrbac.NewAction(req.Action)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid action parameter", slog.String("action", req.Action))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid action"))
		return nil, false
	}

	resource, err := domainrbac.NewResource(req.Resource)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid resource parameter", slog.String("resource", req.Resource))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid resource"))
		return nil, false
	}

	effect, err := domainrbac.NewEffect(req.Effect)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid effect parameter", slog.String("effect", req.Effect))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "effect must be 'allow' or 'deny'"))
		return nil, false
	}

	return &addPolicyParams{
		orgID:    orgID,
		userID:   userID,
		action:   action,
		resource: resource,
		effect:   effect,
	}, true
}

// AddPolicy handles POST /auth/authz/policy.
func (h *AddPolicyHandler) AddPolicy(c *gin.Context) {
	ctx := c.Request.Context()

	params, ok := h.parseRequest(ctx, c)
	if !ok {
		return
	}

	if err := h.policyAdder.AddPolicyForUser(ctx, params.orgID, params.userID, params.action, params.resource, params.effect); err != nil {
		h.logger.ErrorContext(ctx, "add policy", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.Status(http.StatusNoContent)
}
