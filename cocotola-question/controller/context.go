package controller

import libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"

// ContextFieldUserID is a Gin context key for storing the authenticated user's ID.
type ContextFieldUserID = libcontroller.ContextFieldUserID

// ContextFieldOrganizationName is a Gin context key for storing the authenticated user's organization name.
type ContextFieldOrganizationName = libcontroller.ContextFieldOrganizationName

// ContextFieldOrganizationID is a Gin context key for storing the resolved organization ID.
type ContextFieldOrganizationID = libcontroller.ContextFieldOrganizationID

// ContextFieldUserLanguage is a Gin context key for storing the authenticated user's preferred language.
type ContextFieldUserLanguage = libcontroller.ContextFieldUserLanguage
