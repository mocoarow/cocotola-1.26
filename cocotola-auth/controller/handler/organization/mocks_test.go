package organization_test

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	mock "github.com/stretchr/testify/mock"
)

// mockOrganizationFinder is a mock for organization.OrganizationFinder.
type mockOrganizationFinder struct {
	mock.Mock
}

func newMockOrganizationFinder(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockOrganizationFinder {
	m := &mockOrganizationFinder{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *mockOrganizationFinder) FindByName(ctx context.Context, name string) (*domain.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}
