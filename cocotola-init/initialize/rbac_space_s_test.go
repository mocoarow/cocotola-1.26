//go:build small

package initialize_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

func generateSpaceID(t *testing.T) domain.SpaceID {
	t.Helper()
	id, err := domain.NewSpaceIDV7()
	require.NoError(t, err)
	return id
}

func publicSpaceKeyName() string {
	return domainspace.PublicSpaceKeyName(initialize.OrganizationName)
}

func Test_setupRBACPolicies_shouldStopAndReturnError_whenAddPolicyFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: no AddPolicy call is expected after the first failure
	repo := NewMockRBACRepository(t)
	repo.EXPECT().AddPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errRepository).Once()

	// when
	err := initialize.SetupRBACPolicies(ctx, repo, generateOrganizationID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_assignGroup_shouldReturnError_whenAssignGroupToUserFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockRBACRepository(t)
	repo.EXPECT().AssignGroupToUser(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errRepository)

	// when
	err := initialize.AssignGroup(ctx, repo, generateOrganizationID(t), generateAppUserID(t), domainrbac.GroupAdmin(), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreatePublicSpace_shouldReturnExistingSpace_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	orgID := generateOrganizationID(t)
	existing := domainspace.ReconstructSpace(generateSpaceID(t), orgID, domain.SystemAppUserID(), publicSpaceKeyName(), "Public", domainspace.TypePublic(), false)
	repo := NewMockSpaceRepository(t)
	repo.EXPECT().FindByKeyName(mock.Anything, orgID, publicSpaceKeyName()).Return(existing, nil)

	// when
	space, err := initialize.FindOrCreatePublicSpace(ctx, repo, orgID, domain.SystemAppUserID(), discardLogger())

	// then
	require.NoError(t, err)
	assert.Same(t, existing, space)
}

func Test_findOrCreatePublicSpace_shouldSavePublicSpace_whenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID := generateOrganizationID(t)
	ownerID := generateAppUserID(t)
	repo := NewMockSpaceRepository(t)
	repo.EXPECT().FindByKeyName(mock.Anything, orgID, publicSpaceKeyName()).Return(nil, domain.ErrSpaceNotFound)
	var saved *domainspace.Space
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, s *domainspace.Space) error {
		saved = s
		return nil
	})

	// when
	space, err := initialize.FindOrCreatePublicSpace(ctx, repo, orgID, ownerID, discardLogger())

	// then
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Same(t, saved, space)
	assert.Equal(t, orgID, saved.OrganizationID())
	assert.Equal(t, ownerID, saved.OwnerID())
	assert.Equal(t, publicSpaceKeyName(), saved.KeyName())
	assert.True(t, saved.SpaceType().IsPublic())
}

func Test_findOrCreatePublicSpace_shouldReturnError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: Save must not be called
	repo := NewMockSpaceRepository(t)
	repo.EXPECT().FindByKeyName(mock.Anything, mock.Anything, mock.Anything).Return(nil, errRepository)

	// when
	_, err := initialize.FindOrCreatePublicSpace(ctx, repo, generateOrganizationID(t), generateAppUserID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}

func Test_findOrCreatePublicSpace_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := NewMockSpaceRepository(t)
	repo.EXPECT().FindByKeyName(mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrSpaceNotFound)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errRepository)

	// when
	_, err := initialize.FindOrCreatePublicSpace(ctx, repo, generateOrganizationID(t), generateAppUserID(t), discardLogger())

	// then
	require.ErrorIs(t, err, errRepository)
}
