package main

import (
	"context"
	"fmt"

	authdomain "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	questiondomain "github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// authorizationCheckerAdapter adapts cocotola-auth's AuthorizationChecker
// to cocotola-question's AuthorizationChecker interface.
type authorizationCheckerAdapter struct {
	inner authrbac.AuthorizationChecker
}

func (a *authorizationCheckerAdapter) IsAllowed(ctx context.Context, organizationID int, operatorID int, action questiondomain.Action, resource questiondomain.Resource) (bool, error) {
	authAction, err := authrbac.NewAction(action.Value())
	if err != nil {
		return false, fmt.Errorf("convert action: %w", err)
	}

	authResource, err := authrbac.NewResource(resource.Value())
	if err != nil {
		return false, fmt.Errorf("convert resource: %w", err)
	}

	// TODO(uuidv7-phase2): cocotola-question still uses int IDs internally.
	// Until question is migrated to UUIDs, the adapter passes zero-value VOs
	// which means authz checks via this path are effectively unauthenticated.
	// The HTTP-based AuthServiceAuthorizationChecker is the runtime path that
	// matters in production; this in-process adapter is reserved for the
	// monolithic build (cocotola-app).
	_ = organizationID
	_ = operatorID
	return a.inner.IsAllowed(ctx, authdomain.OrganizationID{}, authdomain.AppUserID{}, authAction, authResource)
}
