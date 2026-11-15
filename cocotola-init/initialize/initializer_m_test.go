//go:build medium

package initialize_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

// A fixed login ID keeps reruns from piling new owners onto the shared
// organization's active user list.
const initializeOwnerLoginID = "initialize-test-owner@example.com"

var errSeedPublicWorkbooks = errors.New("seed public workbooks")

// initializeMu serializes Initialize: every run targets the same well-known
// organization, and its find-or-create steps race when run concurrently.
var initializeMu sync.Mutex

func runInitialize(ctx context.Context, t *testing.T, seeder initialize.PublicWorkbookSeeder) error {
	t.Helper()
	initializeMu.Lock()
	defer initializeMu.Unlock()
	if err := initialize.Initialize(ctx, testDB, seeder, initializeOwnerLoginID, ownerPassword); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return nil
}

func newSucceedingSeeder(t *testing.T) *MockPublicWorkbookSeeder {
	t.Helper()
	seeder := NewMockPublicWorkbookSeeder(t)
	seeder.EXPECT().SeedPublicWorkbooks(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return seeder
}

func Test_Initialize_shouldSucceed_whenRunAgain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	require.NoError(t, runInitialize(ctx, t, newSucceedingSeeder(t)))

	// when
	err := runInitialize(ctx, t, newSucceedingSeeder(t))

	// then
	require.NoError(t, err)
}

func Test_Initialize_shouldAssignAdminGroupToOwner_whenRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// when
	require.NoError(t, runInitialize(ctx, t, newSucceedingSeeder(t)))

	// then
	owner, err := gateway.NewAppUserRepository(testDB).FindByLoginID(ctx, initialize.CocotolaOrganizationID(), domain.LoginID(initializeOwnerLoginID))
	require.NoError(t, err)
	groups, err := newRBACRepository(t).GetGroupsForUser(ctx, initialize.CocotolaOrganizationID(), owner.ID())
	require.NoError(t, err)
	assert.Equal(t, []string{domainrbac.GroupAdmin().Value()}, groups)
}

func Test_Initialize_shouldAssignSystemOwnerGroupToSystemAppUser_whenRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// when
	require.NoError(t, runInitialize(ctx, t, newSucceedingSeeder(t)))

	// then
	groups, err := newRBACRepository(t).GetGroupsForUser(ctx, initialize.CocotolaOrganizationID(), domain.SystemAppUserID())
	require.NoError(t, err)
	assert.Equal(t, []string{domainrbac.GroupSystemOwner().Value()}, groups)
}

func Test_Initialize_shouldSeedPublicWorkbooksIntoPublicSpace_whenRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	seeder := newSucceedingSeeder(t)

	// when
	require.NoError(t, runInitialize(ctx, t, seeder))

	// then
	spaceID, err := initialize.NewPublicSpaceFinder(testDB).FindPublicSpaceID(ctx, initialize.CocotolaOrganizationID())
	require.NoError(t, err)
	seeder.AssertCalled(t, "SeedPublicWorkbooks", mock.Anything, initialize.CocotolaOrganizationIDString, spaceID.String())
}

func Test_Initialize_shouldReturnSeederError_whenSeederFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	seeder := NewMockPublicWorkbookSeeder(t)
	seeder.EXPECT().SeedPublicWorkbooks(mock.Anything, mock.Anything, mock.Anything).Return(errSeedPublicWorkbooks)

	// when
	err := runInitialize(ctx, t, seeder)

	// then
	require.ErrorIs(t, err, errSeedPublicWorkbooks)
}
