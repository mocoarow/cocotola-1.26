package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validGroupArgs() (int, int, string, bool) {
	return 1, 1, "group1", true
}

func Test_NewGroup_shouldReturnGroup_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, name, enabled := validGroupArgs()

	// when
	group, err := domain.NewGroup(id, orgID, name, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, group.ID())
	assert.Equal(t, orgID, group.OrganizationID())
	assert.Equal(t, name, group.Name())
	assert.True(t, group.Enabled())
}

func Test_NewGroup_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, name, enabled := validGroupArgs()

	// when
	_, err := domain.NewGroup(0, orgID, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenIDIsNegative(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, name, enabled := validGroupArgs()

	// when
	_, err := domain.NewGroup(-1, orgID, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, name, enabled := validGroupArgs()

	// when
	_, err := domain.NewGroup(id, 0, name, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()

	// when
	_, err := domain.NewGroup(id, orgID, "", enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldReturnError_whenNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()
	longName := strings.Repeat("a", 256)

	// when
	_, err := domain.NewGroup(id, orgID, longName, enabled)

	// then
	require.Error(t, err)
}

func Test_NewGroup_shouldSucceed_whenNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, enabled := validGroupArgs()
	maxName := strings.Repeat("a", 255)

	// when
	group, err := domain.NewGroup(id, orgID, maxName, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxName, group.Name())
}

func Test_Group_Enable_shouldSetEnabledTrue_whenDisabled(t *testing.T) {
	t.Parallel()

	// given
	group, _ := domain.NewGroup(1, 1, "group1", false)

	// when
	group.Enable()

	// then
	assert.True(t, group.Enabled())
}

func Test_Group_Disable_shouldSetEnabledFalse_whenEnabled(t *testing.T) {
	t.Parallel()

	// given
	group, _ := domain.NewGroup(1, 1, "group1", true)

	// when
	group.Disable()

	// then
	assert.False(t, group.Enabled())
}
