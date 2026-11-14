package study_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

var selectionNow = time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

type classification struct {
	dueIDs      []string
	newIDs      []string
	reviewCount int
	newCount    int
}

func classify(t *testing.T, activeIDs []string, records []domainstudy.Record, practice bool) classification {
	t.Helper()
	dueIDs, newIDs, reviewCount, newCount := studyusecase.ClassifyQuestionIDs(activeIDs, records, selectionNow, practice)
	return classification{
		dueIDs:      append([]string{}, dueIDs...),
		newIDs:      append([]string{}, newIDs...),
		reviewCount: reviewCount,
		newCount:    newCount,
	}
}

func recordDueAt(questionID string, nextDueAt time.Time) domainstudy.Record {
	return *domainstudy.ReconstructRecord(fixtureWorkbookID, questionID, 1, nextDueAt.Add(-24*time.Hour), nextDueAt, 1, 0)
}

func idsWithPrefix(prefix string, n int) []string {
	ids := make([]string, 0, n)
	for i := range n {
		ids = append(ids, prefix+strconv.Itoa(i))
	}
	return ids
}

func Test_mixIDs_shouldTakeNineReviewAndOneNew_whenBothPoolsAreLarge(t *testing.T) {
	t.Parallel()

	// given
	review := idsWithPrefix("r", 20)
	unseen := idsWithPrefix("n", 20)

	// when
	got := studyusecase.MixIDs(review, unseen, 10)

	// then
	assert.Equal(t, append(append([]string{}, review[:9]...), unseen[:1]...), got)
}

func Test_mixIDs_shouldFillWithNew_whenReviewPoolIsShort(t *testing.T) {
	t.Parallel()

	// given
	review := idsWithPrefix("r", 3)
	unseen := idsWithPrefix("n", 20)

	// when
	got := studyusecase.MixIDs(review, unseen, 10)

	// then
	assert.Equal(t, append(append([]string{}, review...), unseen[:7]...), got)
}

func Test_mixIDs_shouldFillWithReview_whenNewPoolIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	review := idsWithPrefix("r", 20)

	// when
	got := studyusecase.MixIDs(review, nil, 10)

	// then
	assert.Equal(t, review[:10], got)
}

func Test_mixIDs_shouldReturnAllIDs_whenLimitIsAtLeastTotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		reviewCount int
		newCount    int
		limit       int
	}{
		{name: "limit equals total", reviewCount: 5, newCount: 5, limit: 10},
		{name: "limit exceeds total", reviewCount: 2, newCount: 1, limit: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			review := idsWithPrefix("r", tt.reviewCount)
			unseen := idsWithPrefix("n", tt.newCount)

			// when
			got := studyusecase.MixIDs(review, unseen, tt.limit)

			// then
			assert.Equal(t, append(append([]string{}, review...), unseen...), got)
		})
	}
}

func Test_mixIDs_shouldReturnEmpty_whenLimitIsZero(t *testing.T) {
	t.Parallel()

	// given
	review := idsWithPrefix("r", 5)
	unseen := idsWithPrefix("n", 5)

	// when
	got := studyusecase.MixIDs(review, unseen, 0)

	// then
	assert.Empty(t, got)
}

func Test_mixIDs_shouldTakeNew_whenLimitIsOneAndBothPoolsHaveIDs(t *testing.T) {
	t.Parallel()

	// given
	review := idsWithPrefix("r", 5)
	unseen := idsWithPrefix("n", 5)

	// when
	got := studyusecase.MixIDs(review, unseen, 1)

	// then
	assert.Equal(t, unseen[:1], got)
}

func Test_classifyQuestionIDs_shouldClassifyAsNew_whenQuestionHasNoRecord(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1"}

	// when
	got := classify(t, activeIDs, nil, false)

	// then
	assert.Equal(t, classification{dueIDs: []string{}, newIDs: []string{"q1"}, reviewCount: 0, newCount: 1}, got)
}

func Test_classifyQuestionIDs_shouldClassifyAsDue_whenNextDueAtEqualsNow(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1"}
	records := []domainstudy.Record{recordDueAt("q1", selectionNow)}

	// when
	got := classify(t, activeIDs, records, false)

	// then
	assert.Equal(t, classification{dueIDs: []string{"q1"}, newIDs: []string{}, reviewCount: 1, newCount: 0}, got)
}

func Test_classifyQuestionIDs_shouldSkipQuestion_whenNextDueAtIsInFutureAndNotPractice(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1"}
	records := []domainstudy.Record{recordDueAt("q1", selectionNow.Add(time.Nanosecond))}

	// when
	got := classify(t, activeIDs, records, false)

	// then
	assert.Equal(t, classification{dueIDs: []string{}, newIDs: []string{}, reviewCount: 0, newCount: 0}, got)
}

func Test_classifyQuestionIDs_shouldClassifyAsDue_whenNextDueAtIsInFutureAndPractice(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1"}
	records := []domainstudy.Record{recordDueAt("q1", selectionNow.Add(24*time.Hour))}

	// when
	got := classify(t, activeIDs, records, true)

	// then
	assert.Equal(t, classification{dueIDs: []string{"q1"}, newIDs: []string{}, reviewCount: 1, newCount: 0}, got)
}

func Test_classifyQuestionIDs_shouldClassifyAsNew_whenQuestionHasNoRecordAndPractice(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1", "q2"}
	records := []domainstudy.Record{recordDueAt("q1", selectionNow.Add(24*time.Hour))}

	// when
	got := classify(t, activeIDs, records, true)

	// then
	assert.Equal(t, classification{dueIDs: []string{"q1"}, newIDs: []string{"q2"}, reviewCount: 1, newCount: 1}, got)
}

func Test_classifyQuestionIDs_shouldIgnoreRecord_whenQuestionIsNotActive(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1"}
	records := []domainstudy.Record{recordDueAt("q2", selectionNow.Add(-time.Hour))}

	// when
	got := classify(t, activeIDs, records, false)

	// then
	assert.Equal(t, classification{dueIDs: []string{}, newIDs: []string{"q1"}, reviewCount: 0, newCount: 1}, got)
}

func Test_classifyQuestionIDs_shouldKeepActiveListOrder_whenRecordsAreInDifferentOrder(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1", "q2", "q3", "q4", "q5"}
	records := []domainstudy.Record{
		recordDueAt("q4", selectionNow.Add(-time.Hour)),
		recordDueAt("q1", selectionNow.Add(-time.Hour)),
	}

	// when
	got := classify(t, activeIDs, records, false)

	// then
	assert.Equal(t, classification{dueIDs: []string{"q1", "q4"}, newIDs: []string{"q2", "q3", "q5"}, reviewCount: 2, newCount: 3}, got)
}

func Test_filterExcluded_shouldReturnAllIDs_whenNoActiveIDIsExcluded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		excludeIDs []string
	}{
		{name: "nil", excludeIDs: nil},
		{name: "empty", excludeIDs: []string{}},
		{name: "not active", excludeIDs: []string{"q9"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			activeIDs := []string{"q1", "q2", "q3"}

			// when
			got := studyusecase.FilterExcluded(activeIDs, tt.excludeIDs)

			// then
			assert.Equal(t, []string{"q1", "q2", "q3"}, got)
		})
	}
}

func Test_filterExcluded_shouldRemoveExcludedIDsAndKeepOrder_whenSomeIDsAreExcluded(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1", "q2", "q3", "q4"}

	// when
	got := studyusecase.FilterExcluded(activeIDs, []string{"q3", "q1"})

	// then
	assert.Equal(t, []string{"q2", "q4"}, got)
}

func Test_filterExcluded_shouldReturnEmpty_whenAllIDsAreExcluded(t *testing.T) {
	t.Parallel()

	// given
	activeIDs := []string{"q1", "q2"}

	// when
	got := studyusecase.FilterExcluded(activeIDs, []string{"q2", "q1"})

	// then
	assert.Empty(t, got)
}
