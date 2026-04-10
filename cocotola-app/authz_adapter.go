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

func (a *authorizationCheckerAdapter) IsAllowed(ctx context.Context, organizationID string, operatorID string, action questiondomain.Action, resource questiondomain.Resource) (bool, error) {
	authAction, err := authrbac.NewAction(action.Value())
	if err != nil {
		return false, fmt.Errorf("convert action: %w", err)
	}

	authResource, err := authrbac.NewResource(resource.Value())
	if err != nil {
		return false, fmt.Errorf("convert resource: %w", err)
	}

	orgID, err := authdomain.ParseOrganizationID(organizationID)
	if err != nil {
		return false, fmt.Errorf("parse organization ID: %w", err)
	}

	userID, err := authdomain.ParseAppUserID(operatorID)
	if err != nil {
		return false, fmt.Errorf("parse operator ID: %w", err)
	}

	allowed, err := a.inner.IsAllowed(ctx, orgID, userID, authAction, authResource)
	if err != nil {
		return false, fmt.Errorf("check authorization: %w", err)
	}

	return allowed, nil
}
