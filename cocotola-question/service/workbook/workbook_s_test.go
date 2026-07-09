//go:build small

package workbook_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

// NewCreateWorkbookInput tests

func Test_NewCreateWorkbookInput_shouldSucceed_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given: all required fields are present and within limits.

	// when
	input, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", "My Workbook", "A description", "private", "en")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Equal(t, "My Workbook", input.Title)
	require.Equal(t, "private", input.Visibility)
	require.Equal(t, "en", input.Language)
}

func Test_NewCreateWorkbookInput_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an empty title, which violates the required constraint.

	// when
	_, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", "", "desc", "private", "en")

	// then
	require.Error(t, err)
}

func Test_NewCreateWorkbookInput_shouldSucceed_whenTitleLengthIsAtMaximum(t *testing.T) {
	t.Parallel()

	// given: a title exactly at the 200-character boundary.
	title := strings.Repeat("a", 200)

	// when
	input, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", title, "", "public", "ja")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Len(t, input.Title, 200)
}

func Test_NewCreateWorkbookInput_shouldReturnError_whenTitleExceedsMaximumLength(t *testing.T) {
	t.Parallel()

	// given: a title one character past the 200-character maximum.
	title := strings.Repeat("a", 201)

	// when
	_, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", title, "", "public", "ja")

	// then
	require.Error(t, err)
}

func Test_NewCreateWorkbookInput_shouldReturnError_whenVisibilityIsInvalid(t *testing.T) {
	t.Parallel()

	// given: a visibility value that is not one of the allowed enums (private, public).

	// when
	_, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", "Title", "", "restricted", "en")

	// then
	require.Error(t, err)
}

func Test_NewCreateWorkbookInput_shouldReturnError_whenLanguageCodeIsInvalid(t *testing.T) {
	t.Parallel()

	// given: a language code of length 1 instead of the required 2 characters.

	// when
	_, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", "Title", "", "private", "e")

	// then
	require.Error(t, err)
}

func Test_NewCreateWorkbookInput_shouldSucceed_whenVisibilityIsPublic(t *testing.T) {
	t.Parallel()

	// given: visibility set to the other allowed enum value.

	// when
	input, err := workbookservice.NewCreateWorkbookInput("op1", "org1", "space1", "Title", "", "public", "en")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Equal(t, "public", input.Visibility)
}

// NewUpdateWorkbookInput tests

func Test_NewUpdateWorkbookInput_shouldSucceed_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given: all required fields are present and within limits.

	// when
	input, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", "Updated Title", "Updated desc", "public", "fr")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Equal(t, "Updated Title", input.Title)
}

func Test_NewUpdateWorkbookInput_shouldReturnError_whenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an empty title.

	// when
	_, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", "", "desc", "private", "en")

	// then
	require.Error(t, err)
}

func Test_NewUpdateWorkbookInput_shouldSucceed_whenTitleLengthIsAtMaximum(t *testing.T) {
	t.Parallel()

	// given: a title exactly at the 200-character boundary.
	title := strings.Repeat("b", 200)

	// when
	input, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", title, "", "private", "en")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Len(t, input.Title, 200)
}

func Test_NewUpdateWorkbookInput_shouldReturnError_whenTitleExceedsMaximumLength(t *testing.T) {
	t.Parallel()

	// given: a title one character past the 200-character maximum.
	title := strings.Repeat("b", 201)

	// when
	_, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", title, "", "private", "en")

	// then
	require.Error(t, err)
}

func Test_NewUpdateWorkbookInput_shouldReturnError_whenVisibilityIsInvalid(t *testing.T) {
	t.Parallel()

	// given: an unrecognised visibility value.

	// when
	_, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", "Title", "", "secret", "en")

	// then
	require.Error(t, err)
}

func Test_NewUpdateWorkbookInput_shouldReturnError_whenLanguageCodeIsInvalid(t *testing.T) {
	t.Parallel()

	// given: a three-character language code (the constraint is exactly 2).

	// when
	_, err := workbookservice.NewUpdateWorkbookInput("op1", "org1", "wb1", "Title", "", "public", "eng")

	// then
	require.Error(t, err)
}

// NewGetWorkbookInput tests

func Test_NewGetWorkbookInput_shouldSucceed_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given: all required fields provided.

	// when
	input, err := workbookservice.NewGetWorkbookInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
}

func Test_NewGetWorkbookInput_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an empty workbook ID.

	// when
	_, err := workbookservice.NewGetWorkbookInput("op1", "org1", "")

	// then
	require.Error(t, err)
}

// NewListWorkbooksInput tests

func Test_NewListWorkbooksInput_shouldSucceed_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given: all required fields provided.

	// when
	input, err := workbookservice.NewListWorkbooksInput("op1", "org1", "space1")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
}

func Test_NewListWorkbooksInput_shouldReturnError_whenSpaceIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an empty space ID.

	// when
	_, err := workbookservice.NewListWorkbooksInput("op1", "org1", "")

	// then
	require.Error(t, err)
}

// NewDeleteWorkbookInput tests

func Test_NewDeleteWorkbookInput_shouldSucceed_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given: all required fields provided.

	// when
	input, err := workbookservice.NewDeleteWorkbookInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	require.NotNil(t, input)
}

func Test_NewDeleteWorkbookInput_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an empty workbook ID.

	// when
	_, err := workbookservice.NewDeleteWorkbookInput("op1", "org1", "")

	// then
	require.Error(t, err)
}
