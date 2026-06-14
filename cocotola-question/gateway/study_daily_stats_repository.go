package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
)

const studyDailyStatsSubCollection = "study_daily_stats"

type studyDailyStatRecord struct {
	Date          string    `firestore:"date"`
	Timezone      string    `firestore:"timezone"`
	AnsweredCount int       `firestore:"answeredCount"`
	CorrectCount  int       `firestore:"correctCount"`
	LastUpdatedAt time.Time `firestore:"lastUpdatedAt"`
}

// StudyDailyStatsRepository persists per-day study activity buckets in
// Firestore for the dashboard contribution graph, streak, and daily goal
// features. Each document is keyed by the user-local YYYY-MM-DD date so a
// user has at most ~365 docs/year, and IncrementToday relies on Firestore's
// atomic Increment transform so concurrent answers don't lose writes.
type StudyDailyStatsRepository struct {
	client *firestore.Client
}

// NewStudyDailyStatsRepository returns a new StudyDailyStatsRepository.
func NewStudyDailyStatsRepository(client *firestore.Client) *StudyDailyStatsRepository {
	return &StudyDailyStatsRepository{client: client}
}

func (r *StudyDailyStatsRepository) dailyStatsCol(userID string) *firestore.CollectionRef {
	return r.client.Collection(usersCollection).Doc(userID).Collection(studyDailyStatsSubCollection)
}

// IncrementToday atomically increments today's answered and (optionally)
// correct counts. The first call for a (user, date) creates the document.
// Subsequent calls increment in place via firestore.Increment(1). The
// timezone is persisted alongside the counts so future migrations can detect
// stat docs written under a previous user-setting timezone.
func (r *StudyDailyStatsRepository) IncrementToday(ctx context.Context, userID string, dateKey string, timezone string, correct bool, now time.Time) error {
	if userID == "" {
		return fmt.Errorf("user id is required: %w", domain.ErrInvalidArgument)
	}
	if !study.IsValidDateKey(dateKey) {
		return fmt.Errorf("date key %q must be YYYY-MM-DD: %w", dateKey, domain.ErrInvalidArgument)
	}
	// The timezone arrives via the X-Local-Timezone HTTP header (user-controlled),
	// so the boundary check below is mandatory: time.LoadLocation("") returns UTC
	// silently, hence the explicit empty-check first; LoadLocation then rejects
	// anything that is not a real IANA Time Zone Database entry ("foo",
	// "Not/AZone", control bytes, ...). Without this guard arbitrary strings
	// would be persisted to Firestore alongside legitimate counts.
	if timezone == "" {
		return fmt.Errorf("timezone is required: %w", domain.ErrInvalidArgument)
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("timezone %q must be a valid IANA name: %w", timezone, domain.ErrInvalidArgument)
	}

	data := map[string]any{
		"date":          dateKey,
		"timezone":      timezone,
		"answeredCount": firestore.Increment(1),
		"lastUpdatedAt": now,
	}
	if correct {
		data["correctCount"] = firestore.Increment(1)
	}

	docRef := r.dailyStatsCol(userID).Doc(dateKey)
	if _, err := docRef.Set(ctx, data, firestore.MergeAll); err != nil {
		return fmt.Errorf("increment daily stat for user %s on %s: %w", userID, dateKey, err)
	}
	return nil
}

// FindRange returns the daily stats for the user whose date falls within the
// inclusive [fromDateKey, toDateKey] interval, ordered ascending by date.
// Missing days are omitted: the caller fills gaps with zero buckets when
// rendering. Both bounds must be canonical YYYY-MM-DD keys; fromDateKey must
// be lexically less than or equal to toDateKey (which equals chronological
// ordering for that format).
func (r *StudyDailyStatsRepository) FindRange(ctx context.Context, userID string, fromDateKey string, toDateKey string) ([]study.DailyStat, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required: %w", domain.ErrInvalidArgument)
	}
	if !study.IsValidDateKey(fromDateKey) {
		return nil, fmt.Errorf("from date key %q must be YYYY-MM-DD: %w", fromDateKey, domain.ErrInvalidArgument)
	}
	if !study.IsValidDateKey(toDateKey) {
		return nil, fmt.Errorf("to date key %q must be YYYY-MM-DD: %w", toDateKey, domain.ErrInvalidArgument)
	}
	if fromDateKey > toDateKey {
		return nil, fmt.Errorf("from date %s must not be after to date %s: %w", fromDateKey, toDateKey, domain.ErrInvalidArgument)
	}

	iter := r.dailyStatsCol(userID).
		Where("date", ">=", fromDateKey).
		Where("date", "<=", toDateKey).
		OrderBy("date", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	stats := make([]study.DailyStat, 0)

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			return nil, fmt.Errorf("iterate daily stats for user %s: %w", userID, err)
		}

		var rec studyDailyStatRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("decode daily stat %s: %w", doc.Ref.ID, err)
		}

		stat, err := study.NewDailyStat(rec.Date, rec.AnsweredCount, rec.CorrectCount)
		if err != nil {
			return nil, fmt.Errorf("reconstruct daily stat %s: %w", doc.Ref.ID, err)
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
