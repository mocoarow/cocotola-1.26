package authz_test

import (
	"context"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	mock "github.com/stretchr/testify/mock"
)

// mockAuthorizationChecker is a mock for authz.AuthorizationChecker.
type mockAuthorizationChecker struct {
	mock.Mock
}

func newMockAuthorizationChecker(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockAuthorizationChecker {
	m := &mockAuthorizationChecker{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *mockAuthorizationChecker) IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error) {
	args := m.Called(ctx, organizationID, operatorID, action, resource)
	return args.Bool(0), args.Error(1)
}
