package controller

import (
	authcontroller "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
)

// ContextFieldUserID is a type alias for the auth controller's context key.
// It allows cocotola-question handlers to read the user ID set by the auth middleware.
type ContextFieldUserID = authcontroller.ContextFieldUserID

// ContextFieldOrganizationName is a type alias for the auth controller's context key.
// It allows cocotola-question handlers to read the organization name set by the auth middleware.
type ContextFieldOrganizationName = authcontroller.ContextFieldOrganizationName

// ContextFieldOrganizationID is a Gin context key for storing the resolved organization ID.
type ContextFieldOrganizationID struct{}
