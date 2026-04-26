package domain

import (
	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
)

// AppName is the application identifier used for logging, tracing, and configuration.
const AppName = "cocotola-question"

// SystemAppUserID is the well-known bootstrap user that internal callers
// (e.g. cocotola-init) impersonate when they hit the /api/v1/internal endpoints.
// The canonical value is defined in cocotola-lib/domain.
const SystemAppUserID = libdomain.SystemAppUserIDString
