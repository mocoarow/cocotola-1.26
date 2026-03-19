package auth

import (
	"context"
	"fmt"

	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

// PasswordAuthenticateQuery verifies user credentials.
type PasswordAuthenticateQuery struct {
	userAuthenticator UserAuthenticator
}

// NewPasswordAuthenticateQuery returns a new PasswordAuthenticateQuery.
func NewPasswordAuthenticateQuery(userAuthenticator UserAuthenticator) *PasswordAuthenticateQuery {
	return &PasswordAuthenticateQuery{
		userAuthenticator: userAuthenticator,
	}
}

// PasswordAuthenticate verifies user credentials and returns user info.
func (q *PasswordAuthenticateQuery) PasswordAuthenticate(ctx context.Context, input *authservice.PasswordAuthenticateInput) (*authservice.PasswordAuthenticateOutput, error) {
	userInfo, err := q.userAuthenticator.Authenticate(ctx, input.LoginID, input.Password, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	output, err := authservice.NewPasswordAuthenticateOutput(userInfo.UserID, userInfo.LoginID, userInfo.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("create authenticate output: %w", err)
	}
	return output, nil
}
