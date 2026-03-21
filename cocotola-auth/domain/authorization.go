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

// ActionCreateOrganization returns the action for creating an organization.
func ActionCreateOrganization() RBACAction { return RBACAction{value: "create_organization"} }

// ActionCreateSpace returns the action for creating a space.
func ActionCreateSpace() RBACAction { return RBACAction{value: "create_space"} }

// ActionViewSpace returns the action for viewing a space.
func ActionViewSpace() RBACAction { return RBACAction{value: "view_space"} }

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

// ResourceSpace returns a resource representing a specific space.
func ResourceSpace(spaceID int) RBACResource {
	return RBACResource{value: fmt.Sprintf("space:%d", spaceID)}
}

// RBACGroup represents a named group for authorization.
type RBACGroup struct {
	value string
}

// NewRBACGroup creates a validated RBACGroup.
func NewRBACGroup(value string) (RBACGroup, error) {
	if value == "" {
		return RBACGroup{}, errors.New("rbac group must not be empty")
	}
	return RBACGroup{value: value}, nil
}

// Value returns the string representation.
func (g RBACGroup) Value() string { return g.value }

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

// RBACPolicyRepository manages RBAC policies and group assignments.
type RBACPolicyRepository interface {
	// Group assignment: assign/revoke a group for a user within an organization.
	AssignGroupToUser(ctx context.Context, organizationID int, userID int, group RBACGroup) error
	RevokeGroupFromUser(ctx context.Context, organizationID int, userID int, group RBACGroup) error

	// Policy management: define what actions a group can perform on resources.
	AddPolicy(ctx context.Context, organizationID int, group RBACGroup, action RBACAction, resource RBACResource, effect RBACEffect) error
	RemovePolicy(ctx context.Context, organizationID int, group RBACGroup, action RBACAction, resource RBACResource, effect RBACEffect) error
}

// GroupFinder retrieves groups assigned to a user within an organization.
type GroupFinder interface {
	GetGroupsForUser(ctx context.Context, organizationID int, userID int) ([]string, error)
}

// LoginDeniedGroups returns groups whose members cannot login.
func LoginDeniedGroups() map[string]bool {
	return map[string]bool{
		"system_admin": true,
		"system_owner": true,
	}
}
