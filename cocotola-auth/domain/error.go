package domain

import "errors"

// ErrUnauthenticated is returned when authentication credentials are invalid or missing.
var ErrUnauthenticated = errors.New("unauthenticated")

// ErrTokenNotFound is returned when a token cannot be found in the store.
var ErrTokenNotFound = errors.New("token not found")

// ErrTokenRevoked is returned when a revoked token is used.
var ErrTokenRevoked = errors.New("token revoked")

// ErrSessionExpired is returned when a session token has exceeded its absolute timeout.
var ErrSessionExpired = errors.New("session expired")

// ErrOrganizationNotFound is returned when an organization cannot be found.
var ErrOrganizationNotFound = errors.New("organization not found")

// ErrAppUserNotFound is returned when an app user cannot be found.
var ErrAppUserNotFound = errors.New("app user not found")

// ErrGroupNotFound is returned when a group cannot be found.
var ErrGroupNotFound = errors.New("group not found")

// ErrActiveUserLimitReached is returned when the organization has reached its active user limit.
var ErrActiveUserLimitReached = errors.New("active user limit reached")

// ErrActiveGroupLimitReached is returned when the organization has reached its active group limit.
var ErrActiveGroupLimitReached = errors.New("active group limit reached")

// ErrCyclicGroupHierarchy is returned when adding a group edge would create a cycle.
var ErrCyclicGroupHierarchy = errors.New("cyclic group hierarchy")

// ErrDuplicateEntry is returned when attempting to add a duplicate entry.
var ErrDuplicateEntry = errors.New("duplicate entry")

// ErrSpaceNotFound is returned when a space cannot be found.
var ErrSpaceNotFound = errors.New("space not found")

// ErrForbidden is returned when the operator does not have permission to perform the action.
var ErrForbidden = errors.New("forbidden")
