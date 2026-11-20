//go:build small

package initialize_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

const (
	smallOwnerLoginID  = "owner@example.com"
	smallOwnerPassword = "owner-password"

	spaciousActiveUserLimit = 5
)

var errRepository = errors.New("repository failure")

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func generateOrganizationID(t *testing.T) domain.OrganizationID {
	t.Helper()
	id, err := domain.NewOrganizationIDV7()
	require.NoError(t, err)
	return id
}

func generateAppUserID(t *testing.T) domain.AppUserID {
	t.Helper()
	id, err := domain.NewAppUserIDV7()
	require.NoError(t, err)
	return id
}

func existingOrganization(t *testing.T, maxActiveUsers int) *domain.Organization {
	t.Helper()
	return domain.ReconstructOrganization(generateOrganizationID(t), initialize.OrganizationName, maxActiveUsers, maxActiveUsers)
}

func existingAppUser(t *testing.T, orgID domain.OrganizationID, loginID string) *domainuser.AppUser {
	t.Helper()
	return domainuser.ReconstructAppUser(generateAppUserID(t), orgID, domain.LoginID(loginID), "", true)
}

func newActiveUserList(t *testing.T, orgID domain.OrganizationID, entries ...domain.AppUserID) *domain.ActiveUserList {
	t.Helper()
	list, err := domain.NewActiveUserList(orgID, entries)
	require.NoError(t, err)
	return list
}

func Test_findOrCreateOrganization_shouldReturnExistingOrganization_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	existing := existingOrganization(t, spaciousActiveUserLimit)
	repo := NewMockOrganizationRepository(t)
	repo.EXPECT().FindByName(mock.Anything, initialize.OrganizationName).Return(existing, nil)

	// when
	org, err := initialize.FindOrCreateOrganization(ctx, repo, discardLogger())

	// then
	require.NoError(t, err)
	assert.Equal(t, existing.ID(), org.ID())
}

func Test_findOrCreateOrganization_shouldSaveCocotolaOrganization_whenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockOrganizationRepository(t)
	repo.EXPECT().FindByName(mock.Anything, initialize.OrganizationName).Return(nil, domain.ErrOrganizationNotFound)
	var saved *domain.Organization
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, org *domain.Organization) error {
		saved = org
		return nil
	})

	// when
	org, err := initialize.FindOrCreateOrganization(ctx, repo, discardLogger())

	// then
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Same(t, saved, org)
	assert.Equal(t, initialize.CocotolaOrganizationID(), saved.ID())
	assert.Equal(t, initialize.OrganizationName, saved.Name())
}

func Test_findOrCreateOrganization_shouldReturnError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	repo := NewMockOrganizationRepository(t)
	repo.EXPECT().FindByName(mock.Anything, mock.Anything).Return(nil, errRepository)

	// when
	_, err := initialize.FindOrCreateOrganization(ctx, repo, discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreateOrganization_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockOrganizationRepository(t)
	repo.EXPECT().FindByName(mock.Anything, mock.Anything).Return(nil, domain.ErrOrganizationNotFound)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errRepository)

	// when
	_, err := initialize.FindOrCreateOrganization(ctx, repo, discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreateOwner_shouldReturnExistingOwnerID_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	orgID := generateOrganizationID(t)
	existing := existingAppUser(t, orgID, smallOwnerLoginID)
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, orgID, domain.LoginID(smallOwnerLoginID)).Return(existing, nil)

	// when
	userID, err := initialize.FindOrCreateOwner(ctx, repo, gateway.NewBcryptHasher(), orgID, smallOwnerLoginID, smallOwnerPassword, discardLogger())

	// then
	require.NoError(t, err)
	assert.Equal(t, existing.ID(), userID)
}

func Test_findOrCreateOwner_shouldSaveEnabledOwnerWithHashedPassword_whenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID := generateOrganizationID(t)
	hasher := gateway.NewBcryptHasher()
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, orgID, domain.LoginID(smallOwnerLoginID)).Return(nil, domain.ErrAppUserNotFound)
	var saved *domainuser.AppUser
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, user *domainuser.AppUser) error {
		saved = user
		return nil
	})

	// when
	userID, err := initialize.FindOrCreateOwner(ctx, repo, hasher, orgID, smallOwnerLoginID, smallOwnerPassword, discardLogger())

	// then
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, saved.ID(), userID)
	assert.Equal(t, orgID, saved.OrganizationID())
	assert.Equal(t, domain.LoginID(smallOwnerLoginID), saved.LoginID())
	assert.True(t, saved.Enabled())
	require.NoError(t, hasher.Compare(saved.HashedPassword(), smallOwnerPassword))
}

func Test_findOrCreateOwner_shouldReturnError_whenPasswordIsTooShort(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	orgID := generateOrganizationID(t)
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrAppUserNotFound)

	// when
	_, err := initialize.FindOrCreateOwner(ctx, repo, gateway.NewBcryptHasher(), orgID, smallOwnerLoginID, strings.Repeat("a", domainuser.MinPasswordLength-1), discardLogger())

	// then
	require.ErrorIs(t, err, domainuser.ErrPasswordTooShort)
}

func Test_findOrCreateOwner_shouldReturnError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, mock.Anything, mock.Anything).Return(nil, errRepository)

	// when
	_, err := initialize.FindOrCreateOwner(ctx, repo, gateway.NewBcryptHasher(), generateOrganizationID(t), smallOwnerLoginID, smallOwnerPassword, discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreateOwner_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrAppUserNotFound)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errRepository)

	// when
	_, err := initialize.FindOrCreateOwner(ctx, repo, gateway.NewBcryptHasher(), generateOrganizationID(t), smallOwnerLoginID, smallOwnerPassword, discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreateGuest_shouldReturnExistingGuestID_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	orgID := generateOrganizationID(t)
	guestLoginID := domainuser.NewGuestLoginID(initialize.OrganizationName)
	existing := existingAppUser(t, orgID, guestLoginID)
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, orgID, domain.LoginID(guestLoginID)).Return(existing, nil)

	// when
	userID, err := initialize.FindOrCreateGuest(ctx, repo, orgID, discardLogger())

	// then
	require.NoError(t, err)
	assert.Equal(t, existing.ID(), userID)
}

func Test_findOrCreateGuest_shouldSaveEnabledGuestWithoutPassword_whenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID := generateOrganizationID(t)
	guestLoginID := domainuser.NewGuestLoginID(initialize.OrganizationName)
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, orgID, domain.LoginID(guestLoginID)).Return(nil, domain.ErrAppUserNotFound)
	var saved *domainuser.AppUser
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, user *domainuser.AppUser) error {
		saved = user
		return nil
	})

	// when
	userID, err := initialize.FindOrCreateGuest(ctx, repo, orgID, discardLogger())

	// then
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, saved.ID(), userID)
	assert.Equal(t, orgID, saved.OrganizationID())
	assert.Equal(t, domain.LoginID(guestLoginID), saved.LoginID())
	assert.True(t, saved.Enabled())
	assert.Empty(t, saved.HashedPassword())
}

func Test_findOrCreateGuest_shouldReturnError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, mock.Anything, mock.Anything).Return(nil, errRepository)

	// when
	_, err := initialize.FindOrCreateGuest(ctx, repo, generateOrganizationID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreateGuest_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockAppUserRepository(t)
	repo.EXPECT().FindByLoginID(mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrAppUserNotFound)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errRepository)

	// when
	_, err := initialize.FindOrCreateGuest(ctx, repo, generateOrganizationID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_addToActiveUserList_shouldSaveListContainingUser_whenUserIsNotInList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := existingOrganization(t, spaciousActiveUserLimit)
	userID := generateAppUserID(t)
	repo := NewMockActiveUserListRepository(t)
	repo.EXPECT().FindByOrganizationID(mock.Anything, org.ID()).Return(newActiveUserList(t, org.ID()), nil)
	var saved *domain.ActiveUserList
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, list *domain.ActiveUserList) error {
		saved = list
		return nil
	})

	// when
	err := initialize.AddToActiveUserList(ctx, repo, org, userID, discardLogger())

	// then
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.True(t, saved.Contains(userID))
}

func Test_addToActiveUserList_shouldNotSave_whenUserIsAlreadyInList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	org := existingOrganization(t, spaciousActiveUserLimit)
	userID := generateAppUserID(t)
	repo := NewMockActiveUserListRepository(t)
	repo.EXPECT().FindByOrganizationID(mock.Anything, org.ID()).Return(newActiveUserList(t, org.ID(), userID), nil)

	// when
	err := initialize.AddToActiveUserList(ctx, repo, org, userID, discardLogger())

	// then
	require.NoError(t, err)
}

func Test_addToActiveUserList_shouldReturnError_whenListIsFull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	org := existingOrganization(t, 1)
	repo := NewMockActiveUserListRepository(t)
	repo.EXPECT().FindByOrganizationID(mock.Anything, org.ID()).Return(newActiveUserList(t, org.ID(), generateAppUserID(t)), nil)

	// when
	err := initialize.AddToActiveUserList(ctx, repo, org, generateAppUserID(t), discardLogger())

	// then
	require.ErrorIs(t, err, domain.ErrActiveUserLimitReached)
}

func Test_addToActiveUserList_shouldReturnError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	org := existingOrganization(t, spaciousActiveUserLimit)
	repo := NewMockActiveUserListRepository(t)
	repo.EXPECT().FindByOrganizationID(mock.Anything, mock.Anything).Return(nil, errRepository)

	// when
	err := initialize.AddToActiveUserList(ctx, repo, org, generateAppUserID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_addToActiveUserList_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := existingOrganization(t, spaciousActiveUserLimit)
	repo := NewMockActiveUserListRepository(t)
	repo.EXPECT().FindByOrganizationID(mock.Anything, mock.Anything).Return(newActiveUserList(t, org.ID()), nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errRepository)

	// when
	err := initialize.AddToActiveUserList(ctx, repo, org, generateAppUserID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}
