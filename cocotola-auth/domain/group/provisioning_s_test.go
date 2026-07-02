package group_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

type stubGroupSaver struct {
	err error
}

func (s *stubGroupSaver) Save(_ context.Context, _ *group.Group) error {
	return s.err
}

func Test_Provision_shouldReturnGroup_whenSaverSucceeds(t *testing.T) {
	t.Parallel()

	// given
	saver := &stubGroupSaver{err: nil}

	// when
	g, err := group.Provision(context.Background(), saver, fixtureOrgID, "group1")

	// then
	require.NoError(t, err)
	assert.False(t, g.ID().IsZero())
	assert.Equal(t, fixtureOrgID, g.OrganizationID())
	assert.Equal(t, "group1", g.Name())
	assert.True(t, g.Enabled())
}

func Test_Provision_shouldReturnError_whenSaverFails(t *testing.T) {
	t.Parallel()

	// given
	saverErr := errors.New("db error")
	saver := &stubGroupSaver{err: saverErr}

	// when
	_, err := group.Provision(context.Background(), saver, fixtureOrgID, "group1")

	// then
	require.ErrorIs(t, err, saverErr)
}

func Test_Provision_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	saver := &stubGroupSaver{}

	// when
	_, err := group.Provision(context.Background(), saver, fixtureOrgID, "")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
