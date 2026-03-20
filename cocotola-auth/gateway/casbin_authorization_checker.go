package gateway

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// CasbinAuthorizationChecker implements domain.AuthorizationChecker using Casbin via RBACRepository.
type CasbinAuthorizationChecker struct {
	rbacRepo *RBACRepository
}

var _ domain.AuthorizationChecker = (*CasbinAuthorizationChecker)(nil)

// NewCasbinAuthorizationChecker creates a new CasbinAuthorizationChecker.
func NewCasbinAuthorizationChecker(rbacRepo *RBACRepository) *CasbinAuthorizationChecker {
	return &CasbinAuthorizationChecker{
		rbacRepo: rbacRepo,
	}
}

// IsAllowed checks whether the operator is allowed to perform the action on the resource.
func (c *CasbinAuthorizationChecker) IsAllowed(_ context.Context, organizationID int, operatorID int, action domain.RBACAction, resource domain.RBACResource) (bool, error) {
	if err := c.rbacRepo.LoadPolicy(); err != nil {
		return false, fmt.Errorf("load policy: %w", err)
	}

	ok, err := c.rbacRepo.Enforce(organizationID, operatorID, action, resource)
	if err != nil {
		return false, fmt.Errorf("enforce: %w", err)
	}

	return ok, nil
}
