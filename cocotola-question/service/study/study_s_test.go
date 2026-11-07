package study_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func newRecordAnswerInputForOrdering(regions [][]string) (*studyservice.RecordAnswerInput, error) {
	input, err := studyservice.NewRecordAnswerInputForOrdering(studyservice.RecordAnswerInputForOrderingParams{
		OperatorID:     "op",
		OrganizationID: "org",
		WorkbookID:     "wb",
		QuestionID:     "q",
		Regions:        regions,
	})
	if err != nil {
		return nil, fmt.Errorf("new record answer input for ordering: %w", err)
	}
	return input, nil
}

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

	// given, when
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

func Test_NewRecordAnswerInputForOrdering_shouldSetRegions_whenRegionsAreWithinLimits(t *testing.T) {
	t.Parallel()

	// given
	regions := [][]string{{"am", "I"}, {"a", "student"}}

	// when
	input, err := newRecordAnswerInputForOrdering(regions)

	// then
	require.NoError(t, err)
	require.NotNil(t, input.Regions)
	require.Equal(t, regions, *input.Regions)
}

func Test_NewRecordAnswerInputForOrdering_shouldSucceed_whenRegionCountIsAtLimit(t *testing.T) {
	t.Parallel()

	// given: one chip per region, so the chip total is also within the limit.
	regions := slices.Repeat([][]string{{"a"}}, domainquestion.MaxOrderingSegments)

	// when
	input, err := newRecordAnswerInputForOrdering(regions)

	// then
	require.NoError(t, err)
	require.Len(t, *input.Regions, domainquestion.MaxOrderingSegments)
}

func Test_NewRecordAnswerInputForOrdering_shouldReturnError_whenRegionCountExceedsLimit(t *testing.T) {
	t.Parallel()

	// given: empty regions, so only the region count is over the limit.
	regions := slices.Repeat([][]string{{}}, domainquestion.MaxOrderingSegments+1)

	// when
	_, err := newRecordAnswerInputForOrdering(regions)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRecordAnswerInputForOrdering_shouldSucceed_whenChipCountIsAtLimit(t *testing.T) {
	t.Parallel()

	// given: the chips are split across two regions.
	half := domainquestion.MaxOrderingSegments / 2
	regions := [][]string{
		slices.Repeat([]string{"a"}, half),
		slices.Repeat([]string{"a"}, domainquestion.MaxOrderingSegments-half),
	}

	// when
	input, err := newRecordAnswerInputForOrdering(regions)

	// then
	require.NoError(t, err)
	require.Len(t, *input.Regions, 2)
}

func Test_NewRecordAnswerInputForOrdering_shouldReturnError_whenChipCountExceedsLimit(t *testing.T) {
	t.Parallel()

	// given: each region is within the limit, but the total is one over.
	half := domainquestion.MaxOrderingSegments / 2
	regions := [][]string{
		slices.Repeat([]string{"a"}, half),
		slices.Repeat([]string{"a"}, domainquestion.MaxOrderingSegments-half+1),
	}

	// when
	_, err := newRecordAnswerInputForOrdering(regions)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRecordAnswerInputForOrdering_shouldSucceed_whenTextByteLengthIsAtLimit(t *testing.T) {
	t.Parallel()

	// given
	atLimit := strings.Repeat("a", domainquestion.MaxOrderingTextLength)

	// when
	input, err := newRecordAnswerInputForOrdering([][]string{{atLimit, "b"}})

	// then
	require.NoError(t, err)
	require.Equal(t, atLimit, (*input.Regions)[0][0])
}

func Test_NewRecordAnswerInputForOrdering_shouldReturnError_whenTextByteLengthExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	overLimit := strings.Repeat("a", domainquestion.MaxOrderingTextLength+1)

	// when
	_, err := newRecordAnswerInputForOrdering([][]string{{overLimit, "b"}})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewRecordAnswerInputForOrdering_shouldReturnError_whenMultibyteTextExceedsByteLimit(t *testing.T) {
	t.Parallel()

	// given: 167 characters of 3 bytes each is 501 bytes, though within the limit in characters.
	overLimit := strings.Repeat("あ", 167)
	require.Len(t, overLimit, domainquestion.MaxOrderingTextLength+1)

	// when
	_, err := newRecordAnswerInputForOrdering([][]string{{overLimit, "b"}})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
