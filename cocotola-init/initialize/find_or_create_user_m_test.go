//go:build medium

package initialize_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
)

const ownerPassword = "owner-password"

const userOrgNamePrefix = "user-"

func newBufferLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func ownerLoginID(orgID domain.OrganizationID) string {
	return "owner-" + orgID.String() + "@example.com"
}

func guestLoginID() string {
	return domainuser.NewGuestLoginID(initialize.OrganizationName)
}

func Test_findOrCreateOwner_shouldNotLogLoginID_whenOwnerIsCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, _ := newOrganization(ctx, t, userOrgNamePrefix)
	loginID := ownerLoginID(orgID)
	var buf bytes.Buffer

	// when
	_, err := initialize.FindOrCreateOwner(ctx, gateway.NewAppUserRepository(testDB), gateway.NewBcryptHasher(), orgID, loginID, ownerPassword, newBufferLogger(&buf))

	// then
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), loginID)
}

func Test_findOrCreateOwner_shouldNotLogLoginID_whenOwnerAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, _ := newOrganization(ctx, t, userOrgNamePrefix)
	loginID := ownerLoginID(orgID)
	repo := gateway.NewAppUserRepository(testDB)
	hasher := gateway.NewBcryptHasher()
	_, err := initialize.FindOrCreateOwner(ctx, repo, hasher, orgID, loginID, ownerPassword, slog.New(slog.DiscardHandler))
	require.NoError(t, err)
	var buf bytes.Buffer

	// when
	_, err = initialize.FindOrCreateOwner(ctx, repo, hasher, orgID, loginID, ownerPassword, newBufferLogger(&buf))

	// then
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), loginID)
}

func Test_findOrCreateGuest_shouldNotLogLoginID_whenGuestIsCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, _ := newOrganization(ctx, t, userOrgNamePrefix)
	var buf bytes.Buffer

	// when
	_, err := initialize.FindOrCreateGuest(ctx, gateway.NewAppUserRepository(testDB), orgID, newBufferLogger(&buf))

	// then
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), guestLoginID())
}

func Test_findOrCreateGuest_shouldNotLogLoginID_whenGuestAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgID, _ := newOrganization(ctx, t, userOrgNamePrefix)
	repo := gateway.NewAppUserRepository(testDB)
	_, err := initialize.FindOrCreateGuest(ctx, repo, orgID, slog.New(slog.DiscardHandler))
	require.NoError(t, err)
	var buf bytes.Buffer

	// when
	_, err = initialize.FindOrCreateGuest(ctx, repo, orgID, newBufferLogger(&buf))

	// then
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), guestLoginID())
}
