package authz

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	authzservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/authz"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// PolicyAdder adds per-user RBAC policies.
type PolicyAdder interface {
	AddPolicy(ctx context.Context, input *authzservice.AddPolicyInput) error
}

// AddPolicyHandler handles POST /auth/authz/policy.
type AddPolicyHandler struct {
	policyAdder PolicyAdder
	logger      *slog.Logger
}

// NewAddPolicyHandler returns a new AddPolicyHandler.
func NewAddPolicyHandler(policyAdder PolicyAdder) *AddPolicyHandler {
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

func (h *AddPolicyHandler) bindRequest(ctx context.Context, c *gin.Context) (*addPolicyRequest, bool) {
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

	return &req, true
}

func (h *AddPolicyHandler) newInput(ctx context.Context, c *gin.Context, req *addPolicyRequest) (*authzservice.AddPolicyInput, bool) {
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

	return authzservice.NewAddPolicyInput(orgID, userID, action, resource, effect), true
}

// AddPolicy handles POST /auth/authz/policy.
func (h *AddPolicyHandler) AddPolicy(c *gin.Context) {
	ctx := c.Request.Context()

	req, ok := h.bindRequest(ctx, c)
	if !ok {
		return
	}

	input, ok := h.newInput(ctx, c, req)
	if !ok {
		return
	}

	if err := h.policyAdder.AddPolicy(ctx, input); err != nil {
		h.logger.ErrorContext(ctx, "add policy", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.Status(http.StatusNoContent)
}
