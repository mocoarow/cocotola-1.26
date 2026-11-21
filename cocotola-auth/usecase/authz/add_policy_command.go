package authz

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	authzservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/authz"
)

type userPolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID domain.OrganizationID, userID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) error
}

// AddPolicyCommand adds a per-user RBAC policy.
type AddPolicyCommand struct {
	policyAdder userPolicyAdder
}

// NewAddPolicyCommand returns a new AddPolicyCommand.
func NewAddPolicyCommand(policyAdder userPolicyAdder) *AddPolicyCommand {
	return &AddPolicyCommand{policyAdder: policyAdder}
}

// AddPolicy adds the policy described by input.
func (c *AddPolicyCommand) AddPolicy(ctx context.Context, input *authzservice.AddPolicyInput) error {
	if err := c.policyAdder.AddPolicyForUser(ctx, input.OrganizationID, input.AppUserID, input.Action, input.Resource, input.Effect); err != nil {
		return fmt.Errorf("add policy for user %s: %w", input.AppUserID, err)
	}

	return nil
}
