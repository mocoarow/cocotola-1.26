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

// ErrAppUserAlreadyLinked is returned when attempting to link a provider to an app user
// that is already linked to a provider.
var ErrAppUserAlreadyLinked = errors.New("app user already linked to a provider")

// ErrAppUserAutoLinkRejected is returned when the supabase exchange flow refuses
// to automatically link a Supabase identity to an existing local account because
// the existing account has a password set (and therefore belongs to a human who
// did not opt into linking from the provider side).
var ErrAppUserAutoLinkRejected = errors.New("auto-link rejected: existing account has a password")

// ErrSupabaseEmailNotVerified is returned when the Supabase JWT does not carry
// email_verified=true. The exchange flow MUST refuse to map such a token to
// any local account to prevent email-spoofing account takeover.
var ErrSupabaseEmailNotVerified = errors.New("supabase email not verified")

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

// ErrAppUserProviderNotFound is returned when an app user provider link cannot be found.
var ErrAppUserProviderNotFound = errors.New("app user provider not found")

// ErrSpaceNotFound is returned when a space cannot be found.
var ErrSpaceNotFound = errors.New("space not found")

// ErrUserSettingNotFound is returned when a user setting cannot be found.
var ErrUserSettingNotFound = errors.New("user setting not found")

// ErrForbidden is returned when the operator does not have permission to perform the action.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidArgument is returned when a required field is missing, empty, or invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrInternal is returned when an unexpected infrastructure error occurs.
var ErrInternal = errors.New("internal error")
