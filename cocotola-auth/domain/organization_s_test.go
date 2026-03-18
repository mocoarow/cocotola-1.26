package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validOrganizationArgs() (int, string, int, int) {
	return 1, "org1", 100, 50
}

func Test_NewOrganization_shouldReturnOrganization_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, name, maxUsers, maxGroups := validOrganizationArgs()

	// when
	org, err := domain.NewOrganization(id, name, maxUsers, maxGroups)

	// then
	assert.NoError(t, err)
	assert.Equal(t, id, org.ID())
	assert.Equal(t, name, org.Name())
	assert.Equal(t, maxUsers, org.MaxActiveUsers())
	assert.Equal(t, maxGroups, org.MaxActiveGroups())
}

func Test_NewOrganization_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, name, maxUsers, maxGroups := validOrganizationArgs()

	// when
	_, err := domain.NewOrganization(0, name, maxUsers, maxGroups)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldReturnError_whenIDIsNegative(t *testing.T) {
	t.Parallel()

	// given
	_, name, maxUsers, maxGroups := validOrganizationArgs()

	// when
	_, err := domain.NewOrganization(-1, name, maxUsers, maxGroups)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldReturnError_whenNameIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, _, maxUsers, maxGroups := validOrganizationArgs()

	// when
	_, err := domain.NewOrganization(id, "", maxUsers, maxGroups)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldReturnError_whenMaxActiveUsersIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, name, _, maxGroups := validOrganizationArgs()

	// when
	_, err := domain.NewOrganization(id, name, 0, maxGroups)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldReturnError_whenMaxActiveGroupsIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, name, maxUsers, _ := validOrganizationArgs()

	// when
	_, err := domain.NewOrganization(id, name, maxUsers, 0)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldReturnError_whenNameExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, _, maxUsers, maxGroups := validOrganizationArgs()
	longName := strings.Repeat("a", 256)

	// when
	_, err := domain.NewOrganization(id, longName, maxUsers, maxGroups)

	// then
	assert.Error(t, err)
}

func Test_NewOrganization_shouldSucceed_whenNameIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, _, maxUsers, maxGroups := validOrganizationArgs()
	maxName := strings.Repeat("a", 255)

	// when
	org, err := domain.NewOrganization(id, maxName, maxUsers, maxGroups)

	// then
	assert.NoError(t, err)
	assert.Equal(t, maxName, org.Name())
}
