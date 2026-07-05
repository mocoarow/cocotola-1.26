//go:build medium

package gateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func setupAuthenticator(t *testing.T) *gateway.UserAuthenticator {
	t.Helper()
	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	hasher := gateway.NewBcryptHasher()
	return gateway.NewUserAuthenticator(testDB, hasher, rbacRepo)
}

func setupUserWithPassword(ctx context.Context, t *testing.T, orgID domain.OrganizationID, loginID string, password string, enabled bool) domain.AppUserID {
	t.Helper()
	hasher := gateway.NewBcryptHasher()
	hashed, err := hasher.Hash(password)
	require.NoError(t, err)

	uid, err := domain.NewAppUserIDV7()
	require.NoError(t, err)
	user := domainuser.ReconstructAppUser(uid, orgID, domain.LoginID(loginID), hashed, enabled)

	repo := gateway.NewAppUserRepository(testDB)
	require.NoError(t, repo.Save(ctx, user))
	return uid
}

func Test_UserAuthenticator_Authenticate_shouldReturnUserInfo_whenCredentialsAreValid(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-ok-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "ok@example.com"
	password := "correct-password"
	setupUserWithPassword(ctx, t, orgID, loginID, password, true)
	auth := setupAuthenticator(t)

	// when
	info, err := auth.Authenticate(ctx, loginID, password, orgName)

	// then
	require.NoError(t, err)
	assert.Equal(t, loginID, info.LoginID)
	assert.Equal(t, orgName, info.OrganizationName)
}

func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-nouser-org"
	setupOrganization(ctx, t, testDB, orgName)
	auth := setupAuthenticator(t)

	// when
	_, err := auth.Authenticate(ctx, "nobody@example.com", "any-password", orgName)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenOrganizationDoesNotExist(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	auth := setupAuthenticator(t)

	// when
	_, err := auth.Authenticate(ctx, "user@example.com", "password", "no-such-org")

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenPasswordIsWrong(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-wrongpw-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "wrongpw@example.com"
	setupUserWithPassword(ctx, t, orgID, loginID, "correct-password", true)
	auth := setupAuthenticator(t)

	// when
	_, err := auth.Authenticate(ctx, loginID, "wrong-password", orgName)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenUserIsDisabled(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-disabled-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "disabled@example.com"
	password := "password"
	setupUserWithPassword(ctx, t, orgID, loginID, password, false)
	auth := setupAuthenticator(t)

	// when
	_, err := auth.Authenticate(ctx, loginID, password, orgName)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenUserIsInLoginDeniedGroup(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-deniedgroup-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "denied@example.com"
	password := "password"
	userID := setupUserWithPassword(ctx, t, orgID, loginID, password, true)

	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	// "system_admin" is matched bare by IsLoginDenied — no org prefix.
	loginDeniedGroup := mustGroup(t, "system_admin")
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, loginDeniedGroup))

	hasher := gateway.NewBcryptHasher()
	auth := gateway.NewUserAuthenticator(testDB, hasher, rbacRepo)

	// when
	_, err = auth.Authenticate(ctx, loginID, password, orgName)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_UserAuthenticator_Authenticate_shouldReturnUserInfo_whenUserIsInNonDeniedGroup(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-allowgroup-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "allowed@example.com"
	password := "password"
	userID := setupUserWithPassword(ctx, t, orgID, loginID, password, true)

	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	regularGroup := mustGroup(t, fmt.Sprintf("org:%s,members", orgID.String()))
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, regularGroup))

	hasher := gateway.NewBcryptHasher()
	auth := gateway.NewUserAuthenticator(testDB, hasher, rbacRepo)

	// when
	info, err := auth.Authenticate(ctx, loginID, password, orgName)

	// then
	require.NoError(t, err)
	assert.Equal(t, loginID, info.LoginID)
}

// Verify that IsLoginDenied is matched against the group name, not the scoped Casbin string.
func Test_UserAuthenticator_Authenticate_shouldReturnErrUnauthenticated_whenUserIsInSystemOwnerGroup(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()
	orgName := "auth-sysowner-org"
	orgID := setupOrganization(ctx, t, testDB, orgName)
	loginID := "sysowner@example.com"
	password := "password"
	userID := setupUserWithPassword(ctx, t, orgID, loginID, password, true)

	rbacRepo, err := gateway.NewRBACRepository(testDB)
	require.NoError(t, err)
	// "system_owner" is matched bare by IsLoginDenied — no org prefix.
	ownerGroup := mustGroup(t, "system_owner")
	require.NoError(t, rbacRepo.AssignGroupToUser(ctx, orgID, userID, ownerGroup))

	hasher := gateway.NewBcryptHasher()
	auth := gateway.NewUserAuthenticator(testDB, hasher, rbacRepo)

	// when
	_, err = auth.Authenticate(ctx, loginID, password, orgName)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}
