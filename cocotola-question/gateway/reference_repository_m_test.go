package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

func newReference(t *testing.T, userID, workbookID string) *domainreference.WorkbookReference {
	t.Helper()
	ref, err := domainreference.NewWorkbookReference(userID, workbookID, time.Now())
	require.NoError(t, err)
	return ref
}

func Test_ReferenceRepository_Save_shouldCreateDocument_whenReferenceIsNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)
	ref := newReference(t, "user-save-new-"+t.Name(), "wb-001")

	// when
	err := repo.Save(ctx, ref)

	// then
	require.NoError(t, err)
}

func Test_ReferenceRepository_Save_shouldReturnErrDuplicateReference_whenSamePairSavedTwice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)
	ref := newReference(t, "user-save-dup-"+t.Name(), "wb-dup")
	require.NoError(t, repo.Save(ctx, ref))

	// when
	err := repo.Save(ctx, newReference(t, "user-save-dup-"+t.Name(), "wb-dup"))

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateReference)
}

func Test_ReferenceRepository_FindByID_shouldReturnReference_whenDocumentExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)
	userID := "user-findbyid-" + t.Name()
	workbookID := "wb-findbyid"
	ref := newReference(t, userID, workbookID)
	require.NoError(t, repo.Save(ctx, ref))

	// when
	got, err := repo.FindByID(ctx, userID, workbookID)

	// then
	require.NoError(t, err)
	assert.Equal(t, userID, got.UserID())
	assert.Equal(t, workbookID, got.WorkbookID())
}

func Test_ReferenceRepository_FindByID_shouldReturnErrReferenceNotFound_whenDocumentMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)

	// when
	_, err := repo.FindByID(ctx, "no-such-user", "no-such-wb")

	// then
	require.ErrorIs(t, err, domain.ErrReferenceNotFound)
}

func Test_ReferenceRepository_FindByUserID_shouldReturnAllReferences_whenMultipleExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)
	userID := "user-findbyuserid-" + t.Name()
	for _, wbID := range []string{"wb-a", "wb-b", "wb-c"} {
		require.NoError(t, repo.Save(ctx, newReference(t, userID, wbID)))
	}

	// when
	refs, err := repo.FindByUserID(ctx, userID)

	// then
	require.NoError(t, err)
	assert.Len(t, refs, 3)
}

func Test_ReferenceRepository_Delete_shouldRemoveDocument_whenDocumentExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)
	userID := "user-delete-" + t.Name()
	workbookID := "wb-delete"
	require.NoError(t, repo.Save(ctx, newReference(t, userID, workbookID)))

	// when
	err := repo.Delete(ctx, userID, workbookID)

	// then
	require.NoError(t, err)
	_, err = repo.FindByID(ctx, userID, workbookID)
	require.ErrorIs(t, err, domain.ErrReferenceNotFound)
}

func Test_ReferenceRepository_Delete_shouldSucceed_whenDocumentDoesNotExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewReferenceRepository(client)

	// when
	err := repo.Delete(ctx, "no-such-user", "no-such-wb")

	// then
	require.NoError(t, err)
}
