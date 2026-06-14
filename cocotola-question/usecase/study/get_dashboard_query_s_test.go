package study_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
)

func mustDailyStat(t *testing.T, date string, answered, correct int) domainstudy.DailyStat {
	t.Helper()
	s, err := domainstudy.NewDailyStat(date, answered, correct)
	require.NoError(t, err)
	return s
}

func Test_GetDashboardQuery_shouldReturnZeroOutput_whenNoStats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, "2026-06-08", out.From)
	assert.Equal(t, "2026-06-14", out.To)
	require.Len(t, out.Days, 7)
	assert.Equal(t, 0, out.CurrentStreak)
	assert.Equal(t, 0, out.LongestStreak)
	assert.Equal(t, 0, out.TodayCount)
	assert.Equal(t, 0, out.ActiveDays)
	assert.Equal(t, 0, out.TotalAnswered)
}

func Test_GetDashboardQuery_shouldFillGapsWithZeroBuckets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-10", 5, 4),
		mustDailyStat(t, "2026-06-14", 3, 2),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	require.Len(t, out.Days, 7)
	assert.Equal(t, "2026-06-08", out.Days[0].Date)
	assert.Equal(t, 0, out.Days[0].AnsweredCount)
	assert.Equal(t, "2026-06-10", out.Days[2].Date)
	assert.Equal(t, 5, out.Days[2].AnsweredCount)
	assert.Equal(t, 4, out.Days[2].CorrectCount)
	assert.Equal(t, "2026-06-14", out.Days[6].Date)
	assert.Equal(t, 3, out.Days[6].AnsweredCount)
	assert.Equal(t, 2, out.ActiveDays)
	assert.Equal(t, 8, out.TotalAnswered)
	assert.Equal(t, 6, out.TotalCorrect)
}

func Test_GetDashboardQuery_shouldReturnTodayProgress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-14", 7, 6),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 7, out.TodayCount)
	assert.Equal(t, 6, out.TodayCorrect)
}

func Test_GetDashboardQuery_shouldComputeCurrentStreak_whenTodayHasActivity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-12", 1, 1),
		mustDailyStat(t, "2026-06-13", 1, 1),
		mustDailyStat(t, "2026-06-14", 1, 1),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, out.CurrentStreak)
	assert.Equal(t, 3, out.LongestStreak)
}

func Test_GetDashboardQuery_shouldGracePeriodCurrentStreak_whenTodayEmptyButYesterdayActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-12", 1, 1),
		mustDailyStat(t, "2026-06-13", 1, 1),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, out.CurrentStreak)
}

func Test_GetDashboardQuery_shouldReturnZeroCurrentStreak_whenBothTodayAndYesterdayEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-08", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-10", 5, 5),
		mustDailyStat(t, "2026-06-11", 5, 5),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, out.CurrentStreak)
	assert.Equal(t, 2, out.LongestStreak)
}

func Test_GetDashboardQuery_shouldReportLongestStreak_acrossMultipleRuns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	finder.On("FindRange", mock.Anything, fixtureOperatorID, "2026-06-01", "2026-06-14").Return([]domainstudy.DailyStat{
		mustDailyStat(t, "2026-06-01", 1, 1),
		mustDailyStat(t, "2026-06-02", 1, 1),
		mustDailyStat(t, "2026-06-03", 1, 1),
		mustDailyStat(t, "2026-06-04", 1, 1),
		mustDailyStat(t, "2026-06-05", 1, 1),
		mustDailyStat(t, "2026-06-07", 1, 1),
		mustDailyStat(t, "2026-06-13", 1, 1),
		mustDailyStat(t, "2026-06-14", 1, 1),
	}, nil)
	q := studyusecase.NewGetDashboardQuery(finder)
	input, err := studyservice.NewGetDashboardInput(studyservice.GetDashboardInputParams{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           14,
		TodayDateKey:   "2026-06-14",
	})
	require.NoError(t, err)

	// when
	out, err := q.GetDashboard(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 2, out.CurrentStreak)
	assert.Equal(t, 5, out.LongestStreak)
}

func Test_GetDashboardQuery_shouldReturnError_whenTodayDateKeyMalformed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockdailyStatsFinder(t)
	q := studyusecase.NewGetDashboardQuery(finder)
	input := &studyservice.GetDashboardInput{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Days:           7,
		TodayDateKey:   "2026/06/14",
	}

	// when
	_, err := q.GetDashboard(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
