// Package authz provides service-layer types for authorization policy input.
package authz

import (
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
)

// AddPolicyInput holds the parameters for adding a per-user RBAC policy.
type AddPolicyInput struct {
	OrganizationID domain.OrganizationID
	AppUserID      domain.AppUserID
	Action         domainrbac.Action
	Resource       domainrbac.Resource
	Effect         domainrbac.Effect
}

// NewAddPolicyInput creates an AddPolicyInput.
func NewAddPolicyInput(organizationID domain.OrganizationID, appUserID domain.AppUserID, action domainrbac.Action, resource domainrbac.Resource, effect domainrbac.Effect) *AddPolicyInput {
	return &AddPolicyInput{
		OrganizationID: organizationID,
		AppUserID:      appUserID,
		Action:         action,
		Resource:       resource,
		Effect:         effect,
	}
}
