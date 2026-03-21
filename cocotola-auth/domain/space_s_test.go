package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validSpaceArgs() (int, int, int, string, string, domain.SpaceType, bool) {
	return 1, 1, 1, "public@@org", "Public Space", domain.SpaceTypePublic(), false
}

func Test_NewSpace_shouldReturnSpace_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	space, err := domain.NewSpace(id, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, space.ID())
	assert.Equal(t, orgID, space.OrganizationID())
	assert.Equal(t, ownerID, space.OwnerID())
	assert.Equal(t, keyName, space.KeyName())
	assert.Equal(t, name, space.Name())
	assert.True(t, space.SpaceType().IsPublic())
	assert.False(t, space.Deleted())
}

func Test_NewSpace_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(0, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(id, 0, ownerID, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenOwnerIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(id, orgID, 0, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenKeyNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(id, orgID, ownerID, "", name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenKeyNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()
	longKeyName := strings.Repeat("a", 51)

	// when
	_, err := domain.NewSpace(id, orgID, ownerID, longKeyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldSucceed_whenKeyNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()
	maxKeyName := strings.Repeat("a", 50)

	// when
	space, err := domain.NewSpace(id, orgID, ownerID, maxKeyName, name, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxKeyName, space.KeyName())
}

func Test_NewSpace_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(id, orgID, ownerID, keyName, "", spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()
	longName := strings.Repeat("a", 101)

	// when
	_, err := domain.NewSpace(id, orgID, ownerID, keyName, longName, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldSucceed_whenNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()
	maxName := strings.Repeat("a", 100)

	// when
	space, err := domain.NewSpace(id, orgID, ownerID, keyName, maxName, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxName, space.Name())
}

func Test_NewSpace_shouldReturnError_whenSpaceTypeIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, _, deleted := validSpaceArgs()

	// when
	_, err := domain.NewSpace(id, orgID, ownerID, keyName, name, domain.SpaceType{}, deleted)

	// then
	require.Error(t, err)
}

func Test_ReconstructSpace_shouldReturnSpace_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	space := domain.ReconstructSpace(id, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	assert.Equal(t, id, space.ID())
	assert.Equal(t, orgID, space.OrganizationID())
	assert.Equal(t, ownerID, space.OwnerID())
	assert.Equal(t, keyName, space.KeyName())
	assert.Equal(t, name, space.Name())
	assert.True(t, space.SpaceType().IsPublic())
	assert.False(t, space.Deleted())
}

func Test_Space_Delete_shouldSetDeletedTrue(t *testing.T) {
	t.Parallel()

	// given
	space, _ := domain.NewSpace(1, 1, 1, "key", "name", domain.SpaceTypePublic(), false)

	// when
	space.Delete()

	// then
	assert.True(t, space.Deleted())
}

func Test_Space_Restore_shouldSetDeletedFalse(t *testing.T) {
	t.Parallel()

	// given
	space, _ := domain.NewSpace(1, 1, 1, "key", "name", domain.SpaceTypePublic(), true)

	// when
	space.Restore()

	// then
	assert.False(t, space.Deleted())
}

func Test_PublicSpaceKeyName_shouldReturnPrefixedOrgName(t *testing.T) {
	t.Parallel()

	// when
	keyName := domain.PublicSpaceKeyName("cocotola")

	// then
	assert.Equal(t, "public@@cocotola", keyName)
}

func Test_PrivateSpaceKeyName_shouldReturnPrefixedLoginID(t *testing.T) {
	t.Parallel()

	// when
	keyName := domain.PrivateSpaceKeyName("user@example.com")

	// then
	assert.Equal(t, "private@@user@example.com", keyName)
}
