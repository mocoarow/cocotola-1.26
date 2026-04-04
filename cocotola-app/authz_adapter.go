package main

import (
	"context"
	"fmt"

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

	return a.inner.IsAllowed(ctx, organizationID, operatorID, authAction, authResource)
}
