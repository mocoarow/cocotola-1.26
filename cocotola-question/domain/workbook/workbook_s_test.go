package workbook_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

func validWorkbookArgs() (string, string, string, string, string, string, workbook.Visibility, workbook.Language, time.Time, time.Time) {
	now := time.Now()
	return "wb-1", "space-1", "user-1", "org-1", "Test Workbook", "A test workbook", workbook.VisibilityPrivate(), workbook.LanguageJa(), now, now
}

func Test_NewWorkbook_shouldReturnWorkbook_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, wb.ID())
	assert.Equal(t, spaceID, wb.SpaceID())
	assert.Equal(t, ownerID, wb.OwnerID())
	assert.Equal(t, orgID, wb.OrganizationID())
	assert.Equal(t, title, wb.Title())
	assert.Equal(t, desc, wb.Description())
	assert.True(t, wb.Visibility().IsPrivate())
	assert.Equal(t, "ja", wb.Language().Value())
	assert.Equal(t, 0, wb.Version())
	assert.Equal(t, createdAt, wb.CreatedAt())
	assert.Equal(t, updatedAt, wb.UpdatedAt())
}

func Test_NewWorkbook_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook("", spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenSpaceIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, _, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, "", ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenOwnerIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, _, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, "", orgID, title, desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenOrganizationIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, _, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, "", title, desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, "", desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenTitleExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	longTitle := strings.Repeat("a", 201)

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, longTitle, desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldSucceed_whenTitleIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	maxTitle := strings.Repeat("a", 200)

	// when
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, maxTitle, desc, vis, lang, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, maxTitle, wb.Title())
}

func Test_NewWorkbook_shouldReturnError_whenDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, _, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	longDesc := strings.Repeat("a", 1001)

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, longDesc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenVisibilityIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, _, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, workbook.Visibility{}, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldReturnError_whenLanguageIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, _, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, workbook.Language{}, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ReconstructWorkbook_shouldReturnWorkbook_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	wb := workbook.ReconstructWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, 3, createdAt, updatedAt)

	// then
	assert.Equal(t, id, wb.ID())
	assert.Equal(t, title, wb.Title())
	assert.Equal(t, "ja", wb.Language().Value())
	assert.Equal(t, 3, wb.Version())
}

func Test_Workbook_ChangeVisibility_shouldUpdateVisibility(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, _, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, workbook.VisibilityPrivate(), lang, createdAt, updatedAt)

	// when
	wb.ChangeVisibility(workbook.VisibilityPublic())

	// then
	assert.True(t, wb.Visibility().IsPublic())
}

func Test_Workbook_ChangeLanguage_shouldUpdateLanguage(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, _, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, workbook.LanguageJa(), createdAt, updatedAt)

	// when
	wb.ChangeLanguage(workbook.LanguageEn())

	// then
	assert.Equal(t, "en", wb.Language().Value())
}

func Test_Workbook_UpdateTitle_shouldUpdateTitle(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// when
	err := wb.UpdateTitle("New Title")

	// then
	require.NoError(t, err)
	assert.Equal(t, "New Title", wb.Title())
}

func Test_Workbook_UpdateTitle_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// when
	err := wb.UpdateTitle("")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Workbook_UpdateDescription_shouldUpdateDescription(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// when
	err := wb.UpdateDescription("New Description")

	// then
	require.NoError(t, err)
	assert.Equal(t, "New Description", wb.Description())
}

func Test_Workbook_UpdateDescription_shouldReturnError_whenDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, _ := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)

	// when
	err := wb.UpdateDescription(strings.Repeat("a", 1001))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Workbook_SetVersion_shouldUpdateVersion(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)
	require.NoError(t, err)

	// when
	wb.SetVersion(5)

	// then
	assert.Equal(t, 5, wb.Version())
}

func Test_Workbook_Touch_shouldUpdateUpdatedAt(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)
	require.NoError(t, err)
	newUpdatedAt := updatedAt.Add(time.Hour)

	// when
	wb.Touch(newUpdatedAt)

	// then
	assert.Equal(t, newUpdatedAt, wb.UpdatedAt())
	assert.Equal(t, createdAt, wb.CreatedAt())
}
