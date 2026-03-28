package gateway

import (
	"context"
	"fmt"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
)

// CasbinAuthorizationChecker implements domainrbac.AuthorizationChecker using Casbin via RBACRepository.
type CasbinAuthorizationChecker struct {
	rbacRepo *RBACRepository
}

var _ domainrbac.AuthorizationChecker = (*CasbinAuthorizationChecker)(nil)

// NewCasbinAuthorizationChecker creates a new CasbinAuthorizationChecker.
func NewCasbinAuthorizationChecker(rbacRepo *RBACRepository) *CasbinAuthorizationChecker {
	return &CasbinAuthorizationChecker{
		rbacRepo: rbacRepo,
	}
}

// IsAllowed checks whether the operator is allowed to perform the action on the resource.
func (c *CasbinAuthorizationChecker) IsAllowed(_ context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error) {
	if err := c.rbacRepo.LoadPolicy(); err != nil {
		return false, fmt.Errorf("load policy: %w", err)
	}

	ok, err := c.rbacRepo.Enforce(organizationID, operatorID, action, resource)
	if err != nil {
		return false, fmt.Errorf("enforce: %w", err)
	}

	return ok, nil
}
