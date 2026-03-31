package workbook_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

func validWorkbookArgs() (string, int, int, int, string, string, workbook.Visibility, time.Time, time.Time) {
	now := time.Now()
	return "wb-1", 1, 1, 1, "Test Workbook", "A test workbook", workbook.VisibilityPrivate(), now, now
}

func Test_NewWorkbook_shouldReturnWorkbook_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, wb.ID())
	assert.Equal(t, spaceID, wb.SpaceID())
	assert.Equal(t, ownerID, wb.OwnerID())
	assert.Equal(t, orgID, wb.OrganizationID())
	assert.Equal(t, title, wb.Title())
	assert.Equal(t, desc, wb.Description())
	assert.True(t, wb.Visibility().IsPrivate())
	assert.Equal(t, createdAt, wb.CreatedAt())
	assert.Equal(t, updatedAt, wb.UpdatedAt())
}

func Test_NewWorkbook_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook("", spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenSpaceIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, 0, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenOwnerIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, _, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, 0, orgID, title, desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, _, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, 0, title, desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, "", desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenTitleExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	longTitle := strings.Repeat("a", 201)

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, longTitle, desc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldSucceed_whenTitleIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	maxTitle := strings.Repeat("a", 200)

	// when
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, maxTitle, desc, vis, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxTitle, wb.Title())
}

func Test_NewWorkbook_shouldReturnError_whenDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, _, vis, createdAt, updatedAt := validWorkbookArgs()
	longDesc := strings.Repeat("a", 1001)

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, longDesc, vis, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenVisibilityIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, _, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, workbook.Visibility{}, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_ReconstructWorkbook_shouldReturnWorkbook_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()

	// when
	wb := workbook.ReconstructWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// then
	assert.Equal(t, id, wb.ID())
	assert.Equal(t, title, wb.Title())
}

func Test_Workbook_ChangeVisibility_shouldUpdateVisibility(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, _, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, workbook.VisibilityPrivate(), createdAt, updatedAt)

	// when
	wb.ChangeVisibility(workbook.VisibilityPublic())

	// then
	assert.True(t, wb.Visibility().IsPublic())
}

func Test_Workbook_UpdateTitle_shouldUpdateTitle(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// when
	err := wb.UpdateTitle("New Title")

	// then
	require.NoError(t, err)
	assert.Equal(t, "New Title", wb.Title())
}

func Test_Workbook_UpdateTitle_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// when
	err := wb.UpdateTitle("")

	// then
	require.Error(t, err)
}

func Test_Workbook_UpdateDescription_shouldUpdateDescription(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// when
	err := wb.UpdateDescription("New Description")

	// then
	require.NoError(t, err)
	assert.Equal(t, "New Description", wb.Description())
}

func Test_Workbook_UpdateDescription_shouldReturnError_whenDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, createdAt, updatedAt)

	// when
	err := wb.UpdateDescription(strings.Repeat("a", 1001))

	// then
	require.Error(t, err)
}
