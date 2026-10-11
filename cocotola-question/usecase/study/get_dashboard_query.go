package study

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

const dateLayout = "2006-01-02"

// GetDashboardQuery returns the user-scoped (not workbook-scoped) study dashboard data.
type GetDashboardQuery struct {
	dailyStatsFinder dailyStatsFinder
}

// NewGetDashboardQuery returns a new GetDashboardQuery.
func NewGetDashboardQuery(dailyStatsFinder dailyStatsFinder) *GetDashboardQuery {
	return &GetDashboardQuery{
		dailyStatsFinder: dailyStatsFinder,
	}
}

// GetDashboard returns the dashboard data for the operator.
func (q *GetDashboardQuery) GetDashboard(ctx context.Context, input *studyservice.GetDashboardInput) (*studyservice.GetDashboardOutput, error) {
	if !domainstudy.IsValidDateKey(input.TodayDateKey) {
		return nil, fmt.Errorf("today date key %q must be YYYY-MM-DD: %w", input.TodayDateKey, domain.ErrInvalidArgument)
	}

	todayDate, err := time.Parse(dateLayout, input.TodayDateKey)
	if err != nil {
		return nil, fmt.Errorf("parse today date %q: %w", input.TodayDateKey, domain.ErrInvalidArgument)
	}

	fromDate := todayDate.AddDate(0, 0, -(input.Days - 1))
	fromKey := fromDate.Format(dateLayout)

	stats, err := q.dailyStatsFinder.FindRange(ctx, input.OperatorID, fromKey, input.TodayDateKey)
	if err != nil {
		return nil, fmt.Errorf("find daily stats: %w", err)
	}

	return buildDashboardOutput(input.TodayDateKey, todayDate, fromDate, input.Days, stats), nil
}

// buildDashboardOutput takes pre-parsed dates so it never swallows a
// time.Parse error into a zero time.
func buildDashboardOutput(todayKey string, today, fromDate time.Time, days int, stats []domainstudy.DailyStat) *studyservice.GetDashboardOutput {
	byDate := indexByDate(stats)
	items, totals := buildDailyItems(fromDate, days, byDate)
	current, longest := computeStreaks(byDate, today)

	todayCount, todayCorrect := 0, 0
	if s, ok := byDate[todayKey]; ok {
		todayCount = s.AnsweredCount
		todayCorrect = s.CorrectCount
	}

	return &studyservice.GetDashboardOutput{
		From:          fromDate.Format(dateLayout),
		To:            todayKey,
		Days:          items,
		CurrentStreak: current,
		LongestStreak: longest,
		TodayCount:    todayCount,
		TodayCorrect:  todayCorrect,
		ActiveDays:    totals.activeDays,
		TotalAnswered: totals.answered,
		TotalCorrect:  totals.correct,
	}
}

func indexByDate(stats []domainstudy.DailyStat) map[string]domainstudy.DailyStat {
	m := make(map[string]domainstudy.DailyStat, len(stats))
	for _, s := range stats {
		m[s.Date] = s
	}

	return m
}

type dashboardTotals struct {
	answered   int
	correct    int
	activeDays int
}

func buildDailyItems(fromDate time.Time, days int, byDate map[string]domainstudy.DailyStat) ([]studyservice.DashboardDailyItem, dashboardTotals) {
	items := make([]studyservice.DashboardDailyItem, 0, days)
	totals := dashboardTotals{answered: 0, correct: 0, activeDays: 0}

	for i := range days {
		dateKey := fromDate.AddDate(0, 0, i).Format(dateLayout)
		s, ok := byDate[dateKey]
		if !ok {
			items = append(items, studyservice.DashboardDailyItem{
				Date:          dateKey,
				AnsweredCount: 0,
				CorrectCount:  0,
			})

			continue
		}

		items = append(items, studyservice.DashboardDailyItem{
			Date:          dateKey,
			AnsweredCount: s.AnsweredCount,
			CorrectCount:  s.CorrectCount,
		})
		totals.answered += s.AnsweredCount
		totals.correct += s.CorrectCount

		if s.AnsweredCount > 0 {
			totals.activeDays++
		}
	}

	return items, totals
}

// computeStreaks returns the current and longest calendar-day streaks.
func computeStreaks(byDate map[string]domainstudy.DailyStat, today time.Time) (int, int) {
	answered := answeredDates(byDate)
	if len(answered) == 0 {
		return 0, 0
	}

	current := currentStreak(answered, today)
	longest := longestStreak(answered)

	return current, longest
}

func answeredDates(byDate map[string]domainstudy.DailyStat) map[string]bool {
	answered := make(map[string]bool, len(byDate))
	for _, s := range byDate {
		if s.AnsweredCount > 0 {
			answered[s.Date] = true
		}
	}

	return answered
}

// currentStreak counts back from `today` (or yesterday, when today is
// empty — the one-day grace period so the streak doesn't reset before the
// user has had a chance to answer) until the first gap.
func currentStreak(answered map[string]bool, today time.Time) int {
	cursor := today
	if !answered[cursor.Format(dateLayout)] {
		cursor = cursor.AddDate(0, 0, -1)
		if !answered[cursor.Format(dateLayout)] {
			return 0
		}
	}

	streak := 0

	for answered[cursor.Format(dateLayout)] {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}

	return streak
}

// longestStreak scans every answered date and starts a fresh run only when
// the previous calendar day was a gap — that way each contiguous span is
// counted exactly once instead of being re-counted from every member.
func longestStreak(answered map[string]bool) int {
	longest := 0

	for dateKey := range answered {
		date, err := time.Parse(dateLayout, dateKey)
		if err != nil {
			continue
		}

		prev := date.AddDate(0, 0, -1).Format(dateLayout)
		if answered[prev] {
			continue
		}

		run := runLengthFrom(answered, date)
		if run > longest {
			longest = run
		}
	}

	return longest
}

func runLengthFrom(answered map[string]bool, start time.Time) int {
	run := 1
	next := start.AddDate(0, 0, 1)

	for answered[next.Format(dateLayout)] {
		run++
		next = next.AddDate(0, 0, 1)
	}

	return run
}
