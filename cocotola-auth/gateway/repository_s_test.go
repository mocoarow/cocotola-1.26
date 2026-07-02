//go:build small

package gateway_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

// --- TableName tests ---

func Test_organizationRecord_TableName_shouldReturnOrganization(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.OrganizationRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "organization", tableName)
}

func Test_appUserRecord_TableName_shouldReturnAppUser(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.AppUserRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "app_user", tableName)
}

func Test_groupRecord_TableName_shouldReturnGroup(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.GroupRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "group", tableName)
}

func Test_activeUserRecord_TableName_shouldReturnActiveUser(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.ActiveUserRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "active_user", tableName)
}

func Test_activeGroupRecord_TableName_shouldReturnActiveGroup(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.ActiveGroupRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "active_group", tableName)
}

func Test_userNGroupRecord_TableName_shouldReturnUserNGroup(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.UserNGroupRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "user_n_group", tableName)
}

func Test_groupNGroupRecord_TableName_shouldReturnGroupNGroup(t *testing.T) {
	t.Parallel()
	// given
	record := gateway.GroupNGroupRecordForTest{}
	// when
	tableName := record.TableName()
	// then
	assert.Equal(t, "group_n_group", tableName)
}

// --- toXxxDomain conversion tests ---

func Test_toOrganizationDomain_shouldReconstructOrganization_whenRecordIsValid(t *testing.T) {
	t.Parallel()
	// given
	fixtureOrgIDStr := "00000000-0000-7000-8000-000000000010"
	record := &gateway.OrganizationRecordForTest{
		ID:              fixtureOrgIDStr,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Name:            "org1",
		MaxActiveUsers:  100,
		MaxActiveGroups: 50,
	}
	// when
	org, err := gateway.ToOrganizationDomain(record)
	// then
	require.NoError(t, err)
	assert.Equal(t, domain.MustParseOrganizationID(fixtureOrgIDStr), org.ID())
	assert.Equal(t, "org1", org.Name())
	assert.Equal(t, 100, org.MaxActiveUsers())
	assert.Equal(t, 50, org.MaxActiveGroups())
}

func Test_toAppUserDomain_shouldReconstructAppUser_whenRecordIsValid(t *testing.T) {
	t.Parallel()
	// given
	fixtureUserIDStr := "00000000-0000-7000-8000-000000000020"
	fixtureOrgIDStr := "00000000-0000-7000-8000-000000000010"
	record := &gateway.AppUserRecordForTest{
		ID:             fixtureUserIDStr,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		OrganizationID: fixtureOrgIDStr,
		LoginID:        "user@example.com",
		Enabled:        true,
	}
	// when
	user, err := gateway.ToAppUserDomain(record)
	// then
	require.NoError(t, err)
	assert.Equal(t, domain.MustParseAppUserID(fixtureUserIDStr), user.ID())
	assert.Equal(t, domain.MustParseOrganizationID(fixtureOrgIDStr), user.OrganizationID())
	assert.Equal(t, domain.LoginID("user@example.com"), user.LoginID())
	assert.True(t, user.Enabled())
}

func Test_toAppUserDomain_shouldReconstructDisabledAppUser_whenEnabledIsFalse(t *testing.T) {
	t.Parallel()
	// given
	fixtureUserIDStr := "00000000-0000-7000-8000-000000000021"
	fixtureOrgIDStr := "00000000-0000-7000-8000-000000000011"
	record := &gateway.AppUserRecordForTest{
		ID:             fixtureUserIDStr,
		OrganizationID: fixtureOrgIDStr,
		LoginID:        "disabled@example.com",
		Enabled:        false,
	}
	// when
	user, err := gateway.ToAppUserDomain(record)
	// then
	require.NoError(t, err)
	assert.Equal(t, domain.MustParseAppUserID(fixtureUserIDStr), user.ID())
	assert.False(t, user.Enabled())
}

func Test_toAppUserDomain_shouldReturnError_whenIDIsInvalid(t *testing.T) {
	t.Parallel()
	// given
	record := &gateway.AppUserRecordForTest{
		ID:             "not-a-uuid",
		OrganizationID: "00000000-0000-7000-8000-000000000010",
		LoginID:        "user@example.com",
		Enabled:        true,
	}
	// when
	_, err := gateway.ToAppUserDomain(record)
	// then
	assert.Error(t, err)
}

func Test_toGroupDomain_shouldReconstructGroup_whenRecordIsValid(t *testing.T) {
	t.Parallel()
	// given
	fixtureOrgIDStr := "00000000-0000-7000-8000-000000000010"
	fixtureGroupIDStr := "00000000-0000-7000-8000-000000000005"
	record := &gateway.GroupRecordForTest{
		ID:             fixtureGroupIDStr,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		OrganizationID: fixtureOrgIDStr,
		Name:           "admins",
		Enabled:        true,
	}
	// when
	group, err := gateway.ToGroupDomain(record)
	// then
	require.NoError(t, err)
	assert.Equal(t, domain.MustParseGroupID(fixtureGroupIDStr), group.ID())
	assert.Equal(t, domain.MustParseOrganizationID(fixtureOrgIDStr), group.OrganizationID())
	assert.Equal(t, "admins", group.Name())
	assert.True(t, group.Enabled())
}

func Test_toGroupDomain_shouldReconstructDisabledGroup_whenEnabledIsFalse(t *testing.T) {
	t.Parallel()
	// given
	fixtureOrgIDStr := "00000000-0000-7000-8000-000000000010"
	fixtureGroupIDStr := "00000000-0000-7000-8000-000000000006"
	record := &gateway.GroupRecordForTest{
		ID:             fixtureGroupIDStr,
		OrganizationID: fixtureOrgIDStr,
		Name:           "archived",
		Enabled:        false,
	}
	// when
	group, err := gateway.ToGroupDomain(record)
	// then
	require.NoError(t, err)
	assert.Equal(t, domain.MustParseGroupID(fixtureGroupIDStr), group.ID())
	assert.False(t, group.Enabled())
}

func Test_toOrganizationDomain_shouldReturnError_whenIDIsInvalid(t *testing.T) {
	t.Parallel()
	// given
	record := &gateway.OrganizationRecordForTest{
		ID:   "not-a-uuid",
		Name: "org1",
	}
	// when
	_, err := gateway.ToOrganizationDomain(record)
	// then
	assert.Error(t, err)
}

func Test_toGroupDomain_shouldReturnError_whenIDIsInvalid(t *testing.T) {
	t.Parallel()
	// given
	record := &gateway.GroupRecordForTest{
		ID:             "not-a-uuid",
		OrganizationID: "00000000-0000-7000-8000-000000000010",
		Name:           "admins",
		Enabled:        true,
	}
	// when
	_, err := gateway.ToGroupDomain(record)
	// then
	assert.Error(t, err)
}
