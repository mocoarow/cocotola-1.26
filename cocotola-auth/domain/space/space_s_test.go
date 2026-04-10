package space_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

var (
	fixtureOrgID     = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")
	fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
	fixtureSpaceID1  = domain.MustParseSpaceID("00000000-0000-7000-8000-100000000001")
)

func validSpaceArgs() (domain.SpaceID, domain.OrganizationID, domain.AppUserID, string, string, space.Type, bool) {
	return fixtureSpaceID1, fixtureOrgID, fixtureAppUserID, "public@@org", "Public Space", space.TypePublic(), false
}

func Test_NewSpace_shouldReturnSpace_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	s, err := space.NewSpace(id, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, s.ID())
	assert.Equal(t, orgID, s.OrganizationID())
	assert.Equal(t, ownerID, s.OwnerID())
	assert.Equal(t, keyName, s.KeyName())
	assert.Equal(t, name, s.Name())
	assert.True(t, s.SpaceType().IsPublic())
	assert.False(t, s.Deleted())
}

func Test_NewSpace_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(domain.SpaceID{}, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(id, domain.OrganizationID{}, ownerID, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenOwnerIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(id, orgID, domain.AppUserID{}, keyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenKeyNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(id, orgID, ownerID, "", name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenKeyNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()
	longKeyName := strings.Repeat("a", 51)

	// when
	_, err := space.NewSpace(id, orgID, ownerID, longKeyName, name, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldSucceed_whenKeyNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, _, name, spaceType, deleted := validSpaceArgs()
	maxKeyName := strings.Repeat("a", 50)

	// when
	s, err := space.NewSpace(id, orgID, ownerID, maxKeyName, name, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxKeyName, s.KeyName())
}

func Test_NewSpace_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(id, orgID, ownerID, keyName, "", spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldReturnError_whenNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()
	longName := strings.Repeat("a", 101)

	// when
	_, err := space.NewSpace(id, orgID, ownerID, keyName, longName, spaceType, deleted)

	// then
	require.Error(t, err)
}

func Test_NewSpace_shouldSucceed_whenNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, _, spaceType, deleted := validSpaceArgs()
	maxName := strings.Repeat("a", 100)

	// when
	s, err := space.NewSpace(id, orgID, ownerID, keyName, maxName, spaceType, deleted)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxName, s.Name())
}

func Test_NewSpace_shouldReturnError_whenSpaceTypeIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, _, deleted := validSpaceArgs()

	// when
	_, err := space.NewSpace(id, orgID, ownerID, keyName, name, space.Type{}, deleted)

	// then
	require.Error(t, err)
}

func Test_ReconstructSpace_shouldReturnSpace_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, ownerID, keyName, name, spaceType, deleted := validSpaceArgs()

	// when
	s := space.ReconstructSpace(id, orgID, ownerID, keyName, name, spaceType, deleted)

	// then
	assert.Equal(t, id, s.ID())
	assert.Equal(t, orgID, s.OrganizationID())
	assert.Equal(t, ownerID, s.OwnerID())
	assert.Equal(t, keyName, s.KeyName())
	assert.Equal(t, name, s.Name())
	assert.True(t, s.SpaceType().IsPublic())
	assert.False(t, s.Deleted())
}

func Test_Space_Delete_shouldSetDeletedTrue(t *testing.T) {
	t.Parallel()

	// given
	s, _ := space.NewSpace(fixtureSpaceID1, fixtureOrgID, fixtureAppUserID, "key", "name", space.TypePublic(), false)

	// when
	s.Delete()

	// then
	assert.True(t, s.Deleted())
}

func Test_Space_Restore_shouldSetDeletedFalse(t *testing.T) {
	t.Parallel()

	// given
	s, _ := space.NewSpace(fixtureSpaceID1, fixtureOrgID, fixtureAppUserID, "key", "name", space.TypePublic(), true)

	// when
	s.Restore()

	// then
	assert.False(t, s.Deleted())
}

func Test_PublicSpaceKeyName_shouldReturnPrefixedOrgName(t *testing.T) {
	t.Parallel()

	// when
	keyName := space.PublicSpaceKeyName("cocotola")

	// then
	assert.Equal(t, "public@@cocotola", keyName)
}

func Test_PrivateSpaceKeyName_shouldReturnPrefixedLoginID(t *testing.T) {
	t.Parallel()

	// when
	keyName := space.PrivateSpaceKeyName("user@example.com")

	// then
	assert.Equal(t, "private@@user@example.com", keyName)
}
