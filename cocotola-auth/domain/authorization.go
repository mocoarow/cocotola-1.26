package domain

import (
	"context"
	"errors"
	"fmt"
)

// RBACAction represents an operation type for authorization.
type RBACAction struct {
	value string
}

// NewRBACAction creates a validated RBACAction.
func NewRBACAction(value string) (RBACAction, error) {
	if value == "" {
		return RBACAction{}, errors.New("rbac action must not be empty")
	}
	return RBACAction{value: value}, nil
}

// Value returns the string representation.
func (a RBACAction) Value() string { return a.value }

// ActionCreateUser returns the action for creating a user.
func ActionCreateUser() RBACAction { return RBACAction{value: "create_user"} }

// ActionViewUser returns the action for viewing a user.
func ActionViewUser() RBACAction { return RBACAction{value: "view_user"} }

// ActionDisableUser returns the action for disabling a user.
func ActionDisableUser() RBACAction { return RBACAction{value: "disable_user"} }

// ActionChangePassword returns the action for changing a password.
func ActionChangePassword() RBACAction { return RBACAction{value: "change_password"} }

// ActionCreateGroup returns the action for creating a group.
func ActionCreateGroup() RBACAction { return RBACAction{value: "create_group"} }

// ActionViewGroup returns the action for viewing a group.
func ActionViewGroup() RBACAction { return RBACAction{value: "view_group"} }

// ActionDisableGroup returns the action for disabling a group.
func ActionDisableGroup() RBACAction { return RBACAction{value: "disable_group"} }

// ActionAddUserToGroup returns the action for adding a user to a group.
func ActionAddUserToGroup() RBACAction { return RBACAction{value: "add_user_to_group"} }

// ActionRemoveUserFromGroup returns the action for removing a user from a group.
func ActionRemoveUserFromGroup() RBACAction { return RBACAction{value: "remove_user_from_group"} }

// RBACResource represents a target resource for authorization.
type RBACResource struct {
	value string
}

// NewRBACResource creates a validated RBACResource.
func NewRBACResource(value string) (RBACResource, error) {
	if value == "" {
		return RBACResource{}, errors.New("rbac resource must not be empty")
	}
	return RBACResource{value: value}, nil
}

// Value returns the string representation.
func (r RBACResource) Value() string { return r.value }

// ResourceAny returns a wildcard resource matching all resources.
func ResourceAny() RBACResource { return RBACResource{value: "*"} }

// ResourceUser returns a resource representing a specific user.
func ResourceUser(userID int) RBACResource {
	return RBACResource{value: fmt.Sprintf("user:%d", userID)}
}

// ResourceGroup returns a resource representing a specific group.
func ResourceGroup(groupID int) RBACResource {
	return RBACResource{value: fmt.Sprintf("group:%d", groupID)}
}

// RBACRole represents a named role for authorization.
type RBACRole struct {
	value string
}

// NewRBACRole creates a validated RBACRole.
func NewRBACRole(value string) (RBACRole, error) {
	if value == "" {
		return RBACRole{}, errors.New("rbac role must not be empty")
	}
	return RBACRole{value: value}, nil
}

// Value returns the string representation.
func (r RBACRole) Value() string { return r.value }

// RBACEffect represents the effect of a policy (allow or deny).
type RBACEffect struct {
	value string
}

// Value returns the string representation.
func (e RBACEffect) Value() string { return e.value }

// EffectAllow returns the allow effect.
func EffectAllow() RBACEffect { return RBACEffect{value: "allow"} }

// EffectDeny returns the deny effect.
func EffectDeny() RBACEffect { return RBACEffect{value: "deny"} }

// AuthorizationChecker is a domain service interface for checking permissions.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action RBACAction, resource RBACResource) (bool, error)
}

// RBACPolicyRepository manages RBAC policies and role assignments.
type RBACPolicyRepository interface {
	// Role assignment: assign/revoke a role for a user within an organization.
	AssignRoleToUser(ctx context.Context, organizationID int, userID int, role RBACRole) error
	RevokeRoleFromUser(ctx context.Context, organizationID int, userID int, role RBACRole) error

	// Policy management: define what actions a role can perform on resources.
	AddPolicy(ctx context.Context, organizationID int, role RBACRole, action RBACAction, resource RBACResource, effect RBACEffect) error
	RemovePolicy(ctx context.Context, organizationID int, role RBACRole, action RBACAction, resource RBACResource, effect RBACEffect) error
}
