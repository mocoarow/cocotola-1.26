package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

func newWorkbook(t *testing.T, id, spaceID, ownerID, orgID string) *domainworkbook.Workbook {
	t.Helper()
	now := time.Now()
	wb, err := domainworkbook.NewWorkbook(
		id, spaceID, ownerID, orgID,
		"Test Workbook", "",
		domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(),
		now, now,
	)
	require.NoError(t, err)
	return wb
}

func Test_WorkbookRepository_Save_shouldInsertAndIncrementVersion_whenVersionIsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	wb := newWorkbook(t, "wb-insert-"+t.Name(), "space-1", "owner-1", "org-1")

	// when
	err := repo.Save(ctx, wb)

	// then
	require.NoError(t, err)
	assert.Equal(t, 1, wb.Version())
}

func Test_WorkbookRepository_Save_shouldPersistAggregate_whenInsertSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	id := "wb-persist-" + t.Name()
	wb := newWorkbook(t, id, "space-2", "owner-2", "org-2")
	require.NoError(t, repo.Save(ctx, wb))

	// when
	loaded, err := repo.FindByID(ctx, id)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, loaded.ID())
	assert.Equal(t, "space-2", loaded.SpaceID())
	assert.Equal(t, "owner-2", loaded.OwnerID())
	assert.Equal(t, "org-2", loaded.OrganizationID())
	assert.Equal(t, "Test Workbook", loaded.Title())
	assert.Equal(t, 1, loaded.Version())
}

func Test_WorkbookRepository_Save_shouldUpdateAndBumpVersion_whenVersionMatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	id := "wb-update-" + t.Name()
	wb := newWorkbook(t, id, "space-3", "owner-3", "org-3")
	require.NoError(t, repo.Save(ctx, wb))

	wb.ChangeVisibility(domainworkbook.VisibilityPublic())

	// when
	err := repo.Save(ctx, wb)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, wb.Version())

	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.True(t, loaded.Visibility().IsPublic())
	assert.Equal(t, 2, loaded.Version())
}

func Test_WorkbookRepository_Save_shouldReturnConcurrentModification_whenVersionMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	id := "wb-conflict-" + t.Name()
	wb := newWorkbook(t, id, "space-4", "owner-4", "org-4")
	require.NoError(t, repo.Save(ctx, wb))

	// stale aggregate still at version 0
	stale := newWorkbook(t, id, "space-4", "owner-4", "org-4")

	// when
	err := repo.Save(ctx, stale)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}

func Test_WorkbookRepository_Save_shouldReturnErrWorkbookNotFound_whenDocWasDeletedAfterLoad(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a saved workbook is deleted out-of-band before a stale aggregate tries to save
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	id := "wb-deleted-then-save-" + t.Name()
	wb := newWorkbook(t, id, "space-5", "owner-5", "org-5")
	require.NoError(t, repo.Save(ctx, wb))

	loaded, err := repo.FindByID(ctx, id)
	require.NoError(t, err)

	// delete the underlying document directly
	_, err = client.Collection("workbooks").Doc(id).Delete(ctx)
	require.NoError(t, err)

	// when: the stale loaded aggregate tries to save
	loaded.ChangeVisibility(domainworkbook.VisibilityPublic())
	err = repo.Save(ctx, loaded)

	// then: callers see a domain not-found, not a generic error
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
	assert.NotErrorIs(t, err, libversioned.ErrConcurrentModification,
		"deleted document must surface as not-found, not as concurrent modification")
}

func Test_WorkbookRepository_FindByID_shouldReturnErrWorkbookNotFound_whenMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)

	// when
	_, err := repo.FindByID(ctx, "nonexistent-workbook-"+t.Name())

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}

func Test_WorkbookRepository_FindBySpaceID_shouldReturnMatchingWorkbooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	spaceID := "space-find-by-space-" + t.Name()

	wb1 := newWorkbook(t, "wb-s1-"+t.Name(), spaceID, "owner-1", "org-1")
	wb2 := newWorkbook(t, "wb-s2-"+t.Name(), spaceID, "owner-2", "org-2")
	require.NoError(t, repo.Save(ctx, wb1))
	require.NoError(t, repo.Save(ctx, wb2))

	// when
	got, err := repo.FindBySpaceID(ctx, spaceID)

	// then
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func Test_WorkbookRepository_FindBySpaceID_shouldReturnEmpty_whenNoWorkbooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)

	// when
	got, err := repo.FindBySpaceID(ctx, "nonexistent-space-"+t.Name())

	// then
	require.NoError(t, err)
	assert.Empty(t, got)
}

func Test_WorkbookRepository_FindPublicByOrganizationIDAndLanguage_shouldReturnOnlyPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	orgID := "org-public-filter-" + t.Name()
	spaceID := "space-filter-" + t.Name()
	now := time.Now()

	publicWB, err := domainworkbook.NewWorkbook(
		"wb-public-"+t.Name(), spaceID, "owner-1", orgID,
		"Public WB", "", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(),
		now, now,
	)
	require.NoError(t, err)

	privateWB, err := domainworkbook.NewWorkbook(
		"wb-private-"+t.Name(), spaceID, "owner-2", orgID,
		"Private WB", "", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(),
		now, now,
	)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, publicWB))
	require.NoError(t, repo.Save(ctx, privateWB))

	// when
	got, err := repo.FindPublicByOrganizationIDAndLanguage(ctx, orgID, "ja")

	// then
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, publicWB.ID(), got[0].ID())
	assert.True(t, got[0].Visibility().IsPublic())
}

func Test_WorkbookRepository_FindPublicByOrganizationIDAndLanguage_shouldFilterByLanguage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	orgID := "org-lang-filter-" + t.Name()
	spaceID := "space-lang-filter-" + t.Name()
	now := time.Now()

	jaWB, err := domainworkbook.NewWorkbook(
		"wb-ja-"+t.Name(), spaceID, "owner-1", orgID,
		"Japanese WB", "", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(),
		now, now,
	)
	require.NoError(t, err)

	enWB, err := domainworkbook.NewWorkbook(
		"wb-en-"+t.Name(), spaceID, "owner-2", orgID,
		"English WB", "", domainworkbook.VisibilityPublic(), domainworkbook.LanguageEn(),
		now, now,
	)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, jaWB))
	require.NoError(t, repo.Save(ctx, enWB))

	// when
	got, err := repo.FindPublicByOrganizationIDAndLanguage(ctx, orgID, "ja")

	// then
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, jaWB.ID(), got[0].ID())
}

func Test_WorkbookRepository_Delete_shouldRemoveDocument(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := setupFirestoreClient(t)
	repo := gateway.NewWorkbookRepository(client)
	id := "wb-delete-" + t.Name()
	wb := newWorkbook(t, id, "space-del", "owner-del", "org-del")
	require.NoError(t, repo.Save(ctx, wb))

	// when
	err := repo.Delete(ctx, id)

	// then
	require.NoError(t, err)
	_, err = repo.FindByID(ctx, id)
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
