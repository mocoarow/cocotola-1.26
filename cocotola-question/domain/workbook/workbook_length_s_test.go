package workbook_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

func multibyteText(n int) string {
	return strings.Repeat("あ", n)
}

func newValidWorkbook(t *testing.T) *workbook.Workbook {
	t.Helper()

	id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()
	wb, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, desc, vis, lang, createdAt, updatedAt)
	require.NoError(t, err)
	return wb
}

func Test_NewWorkbook_shouldSucceed_whenMultibyteTitleIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, multibyteText(workbook.MaxTitleLength), desc, vis, lang, createdAt, updatedAt)

	// then
	require.NoError(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenMultibyteTitleExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, _, desc, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, multibyteText(workbook.MaxTitleLength+1), desc, vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewWorkbook_shouldSucceed_whenMultibyteDescriptionIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, _, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, multibyteText(workbook.MaxDescriptionLength), vis, lang, createdAt, updatedAt)

	// then
	require.NoError(t, err)
}

func Test_NewWorkbook_shouldReturnError_whenMultibyteDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, spaceID, ownerID, orgID, title, _, vis, lang, createdAt, updatedAt := validWorkbookArgs()

	// when
	_, err := workbook.NewWorkbook(id, spaceID, ownerID, orgID, title, multibyteText(workbook.MaxDescriptionLength+1), vis, lang, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Workbook_UpdateTitle_shouldSucceed_whenMultibyteTitleIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	wb := newValidWorkbook(t)

	// when
	err := wb.UpdateTitle(multibyteText(workbook.MaxTitleLength))

	// then
	require.NoError(t, err)
}

func Test_Workbook_UpdateTitle_shouldReturnError_whenMultibyteTitleExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	wb := newValidWorkbook(t)

	// when
	err := wb.UpdateTitle(multibyteText(workbook.MaxTitleLength + 1))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Workbook_UpdateDescription_shouldSucceed_whenMultibyteDescriptionIsAtMaxLength(t *testing.T) {
	t.Parallel()

	// given
	wb := newValidWorkbook(t)

	// when
	err := wb.UpdateDescription(multibyteText(workbook.MaxDescriptionLength))

	// then
	require.NoError(t, err)
}

func Test_Workbook_UpdateDescription_shouldReturnError_whenMultibyteDescriptionExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	wb := newValidWorkbook(t)

	// when
	err := wb.UpdateDescription(multibyteText(workbook.MaxDescriptionLength + 1))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
