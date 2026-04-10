package group_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

func validGroupArgs() (int, domain.OrganizationID, string, bool) {
	return 1, fixtureOrgID, "group1", true
}

func Test_NewGroup_shouldReturnGroup_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, name, enabled := validGroupArgs()

	// when
	g, err := group.NewGroup(id, orgID, name, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, g.ID())
	assert.Equal(t, orgID, g.OrganizationID())
	assert.Equal(t, name, g.Name())
	assert.True(t, g.Enabled())
}

func Test_NewGroup_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, name, enabled := validGroupArgs()

	// when
	_, err := group.NewGroup(0, orgID, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenIDIsNegative(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, name, enabled := validGroupArgs()

	// when
	_, err := group.NewGroup(-1, orgID, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, name, enabled := validGroupArgs()

	// when
	_, err := group.NewGroup(id, domain.OrganizationID{}, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()

	// when
	_, err := group.NewGroup(id, orgID, "", enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()
	longName := strings.Repeat("a", 256)

	// when
	_, err := group.NewGroup(id, orgID, longName, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldSucceed_whenNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()
	maxName := strings.Repeat("a", 255)

	// when
	g, err := group.NewGroup(id, orgID, maxName, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxName, g.Name())
}

func Test_Group_Enable_shouldSetEnabledTrue_whenDisabled(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewGroup(1, fixtureOrgID, "group1", false)

	// when
	g.Enable()

	// then
	assert.True(t, g.Enabled())
}

func Test_Group_Disable_shouldSetEnabledFalse_whenEnabled(t *testing.T) {
	t.Parallel()

	// given
	g, _ := group.NewGroup(1, fixtureOrgID, "group1", true)

	// when
	g.Disable()

	// then
	assert.False(t, g.Enabled())
}
