package gateway

// Test helpers - export internal types and conversion functions for testing.

type OrganizationRecordForTest = organizationRecord
type AppUserRecordForTest = appUserRecord
type GroupRecordForTest = groupRecord
type ActiveUserRecordForTest = activeUserRecord
type ActiveGroupRecordForTest = activeGroupRecord
type UserNGroupRecordForTest = userNGroupRecord
type GroupNGroupRecordForTest = groupNGroupRecord
type AppUserProviderRecordForTest = appUserProviderRecord

var ToOrganizationDomain = toOrganizationDomain
var ToAppUserDomain = toAppUserDomain
var ToGroupDomain = toGroupDomain
