package study_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_NewGetStudyQuestionsInput_shouldSucceed_whenExcludeIDsCountIsAtLimit(t *testing.T) {
	t.Parallel()

	// given: an excludeIDs slice exactly at the maximum allowed size.
	excludeIDs := make([]string, studyservice.MaxExcludeIDsCount)
	for i := range excludeIDs {
		excludeIDs[i] = "q"
	}

	// when
	input, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		Limit:          10,
		Practice:       false,
		ExcludeIDs:     excludeIDs,
	})

	// then: the boundary case is accepted (the check uses > not >=).
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Len(t, input.ExcludeIDs, studyservice.MaxExcludeIDsCount)
}

func Test_NewGetStudyQuestionsInput_shouldReturnError_whenExcludeIDsCountExceedsLimit(t *testing.T) {
	t.Parallel()

	// given: an excludeIDs slice one item past the maximum.
	excludeIDs := make([]string, studyservice.MaxExcludeIDsCount+1)
	for i := range excludeIDs {
		excludeIDs[i] = "q"
	}

	// when
	_, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		Limit:          10,
		ExcludeIDs:     excludeIDs,
	})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewGetStudyQuestionsInput_shouldSucceed_whenExcludeIDLengthIsAtLimit(t *testing.T) {
	t.Parallel()

	// given: a single excludeID exactly at the per-element max length.
	atLimit := strings.Repeat("a", studyservice.MaxExcludeIDLength)

	// when
	input, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		Limit:          10,
		ExcludeIDs:     []string{atLimit},
	})

	// then: the boundary case is accepted (the check uses > not >=).
	require.NoError(t, err)
	require.NotNil(t, input)
	require.Equal(t, atLimit, input.ExcludeIDs[0])
}

func Test_NewGetStudyQuestionsInput_shouldReturnError_whenExcludeIDExceedsLength(t *testing.T) {
	t.Parallel()

	// given: a single excludeID one byte longer than the per-element limit.
	overLimit := strings.Repeat("a", studyservice.MaxExcludeIDLength+1)

	// when
	_, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		Limit:          10,
		ExcludeIDs:     []string{overLimit},
	})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewGetStudyQuestionsInput_shouldReturnError_whenExcludeIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given: an excludeIDs slice containing an empty string. Question IDs are
	// never empty in this domain, so a client sending one is malformed and the
	// server (the trust boundary) must reject it independently of any browser
	// filtering.

	// when
	_, err := studyservice.NewGetStudyQuestionsInput(studyservice.GetStudyQuestionsInputParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		Limit:          10,
		ExcludeIDs:     []string{""},
	})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
