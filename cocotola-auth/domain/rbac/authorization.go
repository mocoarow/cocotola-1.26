// Package rbac provides role-based access control types and interfaces.
package rbac

import (
	"context"
	"errors"
	"fmt"
)

// Action represents an operation type for authorization.
type Action struct {
	value string
}

// NewAction creates a validated RBACAction.
func NewAction(value string) (Action, error) {
	if value == "" {
		return Action{}, errors.New("rbac action must not be empty")
	}
	return Action{value: value}, nil
}

// Value returns the string representation.
func (a Action) Value() string { return a.value }

// ActionCreateUser returns the action for creating a user.
func ActionCreateUser() Action { return Action{value: "create_user"} }

// ActionViewUser returns the action for viewing a user.
func ActionViewUser() Action { return Action{value: "view_user"} }

// ActionDisableUser returns the action for disabling a user.
func ActionDisableUser() Action { return Action{value: "disable_user"} }

// ActionChangePassword returns the action for changing a password.
func ActionChangePassword() Action { return Action{value: "change_password"} }

// ActionCreateGroup returns the action for creating a group.
func ActionCreateGroup() Action { return Action{value: "create_group"} }

// ActionViewGroup returns the action for viewing a group.
func ActionViewGroup() Action { return Action{value: "view_group"} }

// ActionDisableGroup returns the action for disabling a group.
func ActionDisableGroup() Action { return Action{value: "disable_group"} }

// ActionAddUserToGroup returns the action for adding a user to a group.
func ActionAddUserToGroup() Action { return Action{value: "add_user_to_group"} }

// ActionRemoveUserFromGroup returns the action for removing a user from a group.
func ActionRemoveUserFromGroup() Action { return Action{value: "remove_user_from_group"} }

// ActionCreateOrganization returns the action for creating an organization.
func ActionCreateOrganization() Action { return Action{value: "create_organization"} }

// ActionCreateSpace returns the action for creating a space.
func ActionCreateSpace() Action { return Action{value: "create_space"} }

// ActionListSpaces returns the action for listing spaces (metadata only).
func ActionListSpaces() Action { return Action{value: "list_spaces"} }

// ActionViewSpace returns the action for viewing a space.
func ActionViewSpace() Action { return Action{value: "view_space"} }

// ActionCreateWorkbook returns the action for creating a workbook.
func ActionCreateWorkbook() Action { return Action{value: "create_workbook"} }

// ActionViewWorkbook returns the action for viewing a workbook.
func ActionViewWorkbook() Action { return Action{value: "view_workbook"} }

// ActionUpdateWorkbook returns the action for updating a workbook.
func ActionUpdateWorkbook() Action { return Action{value: "update_workbook"} }

// ActionDeleteWorkbook returns the action for deleting a workbook.
func ActionDeleteWorkbook() Action { return Action{value: "delete_workbook"} }

// ActionImportWorkbook returns the action for importing (referencing) a workbook.
func ActionImportWorkbook() Action { return Action{value: "import_workbook"} }

// ActionCreateQuestion returns the action for creating a question.
func ActionCreateQuestion() Action { return Action{value: "create_question"} }

// ActionUpdateQuestion returns the action for updating a question.
func ActionUpdateQuestion() Action { return Action{value: "update_question"} }

// ActionDeleteQuestion returns the action for deleting a question.
func ActionDeleteQuestion() Action { return Action{value: "delete_question"} }

// Resource represents a target resource for authorization.
type Resource struct {
	value string
}

// NewResource creates a validated RBACResource.
func NewResource(value string) (Resource, error) {
	if value == "" {
		return Resource{}, errors.New("rbac resource must not be empty")
	}
	return Resource{value: value}, nil
}

// Value returns the string representation.
func (r Resource) Value() string { return r.value }

// ResourceAny returns a wildcard resource matching all resources.
func ResourceAny() Resource { return Resource{value: "*"} }

// ResourceUser returns a resource representing a specific user.
func ResourceUser(userID int) Resource {
	return Resource{value: fmt.Sprintf("user:%d", userID)}
}

// ResourceGroup returns a resource representing a specific group.
func ResourceGroup(groupID int) Resource {
	return Resource{value: fmt.Sprintf("group:%d", groupID)}
}

// ResourceSpace returns a resource representing a specific space.
func ResourceSpace(spaceID int) Resource {
	return Resource{value: fmt.Sprintf("space:%d", spaceID)}
}

// ResourceWorkbook returns a resource representing a specific workbook.
func ResourceWorkbook(workbookID string) Resource {
	return Resource{value: "workbook:" + workbookID}
}

// ResourceQuestion returns a resource representing a specific question.
func ResourceQuestion(questionID string) Resource {
	return Resource{value: "question:" + questionID}
}

// Group represents a named group for authorization.
type Group struct {
	value string
}

// NewGroup creates a validated RBACGroup.
func NewGroup(value string) (Group, error) {
	if value == "" {
		return Group{}, errors.New("rbac group must not be empty")
	}
	return Group{value: value}, nil
}

// Value returns the string representation.
func (g Group) Value() string { return g.value }

// Effect represents the effect of a policy (allow or deny).
type Effect struct {
	value string
}

// Value returns the string representation.
func (e Effect) Value() string { return e.value }

// EffectAllow returns the allow effect.
func EffectAllow() Effect { return Effect{value: "allow"} }

// EffectDeny returns the deny effect.
func EffectDeny() Effect { return Effect{value: "deny"} }

// AuthorizationChecker is a domain service interface for checking permissions.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action Action, resource Resource) (bool, error)
}

// PolicyRepository manages RBAC policies and group assignments.
type PolicyRepository interface {
	// Group assignment: assign/revoke a group for a user within an organization.
	AssignGroupToUser(ctx context.Context, organizationID int, userID int, group Group) error
	RevokeGroupFromUser(ctx context.Context, organizationID int, userID int, group Group) error

	// Policy management: define what actions a group can perform on resources.
	AddPolicy(ctx context.Context, organizationID int, group Group, action Action, resource Resource, effect Effect) error
	RemovePolicy(ctx context.Context, organizationID int, group Group, action Action, resource Resource, effect Effect) error
}

// UserPolicyManager manages per-user RBAC policies.
type UserPolicyManager interface {
	AddPolicyForUser(ctx context.Context, organizationID int, userID int, action Action, resource Resource, effect Effect) error
}

// GroupFinder retrieves groups assigned to a user within an organization.
type GroupFinder interface {
	GetGroupsForUser(ctx context.Context, organizationID int, userID int) ([]string, error)
}

// IsLoginDenied returns true if the given group denies login.
func IsLoginDenied(group string) bool {
	switch group {
	case "system_admin", "system_owner":
		return true
	default:
		return false
	}
}
