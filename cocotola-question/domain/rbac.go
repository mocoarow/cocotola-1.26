package domain

import (
	"context"
	"errors"
)

// ErrEmptyActionValue is returned when an empty action value is provided.
var ErrEmptyActionValue = errors.New("action value must not be empty")

// ErrEmptyResourceValue is returned when an empty resource value is provided.
var ErrEmptyResourceValue = errors.New("resource value must not be empty")

// Action represents an RBAC operation type for authorization.
type Action struct {
	value string
}

// NewAction creates an Action from a string value with validation.
func NewAction(value string) (Action, error) {
	if value == "" {
		return Action{}, ErrEmptyActionValue
	}

	return Action{value: value}, nil
}

// Value returns the string representation.
func (a Action) Value() string { return a.value }

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

// NewResource creates a Resource from a string value with validation.
func NewResource(value string) (Resource, error) {
	if value == "" {
		return Resource{}, ErrEmptyResourceValue
	}

	return Resource{value: value}, nil
}

// Value returns the string representation.
func (r Resource) Value() string { return r.value }

// ResourceAny returns a wildcard resource matching all resources.
func ResourceAny() Resource { return Resource{value: "*"} }

// ResourceWorkbook returns a resource representing a specific workbook.
func ResourceWorkbook(workbookID string) Resource {
	return Resource{value: "workbook:" + workbookID}
}

// AuthorizationChecker checks if an action is allowed by RBAC policy.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action Action, resource Resource) (bool, error)
}

// Effect represents a policy effect (allow or deny).
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// PolicyAdder adds per-user RBAC policies via the auth service.
type PolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID string, userID string, action Action, resource Resource, effect string) error
}
