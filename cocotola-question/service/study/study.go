// Package study provides service-layer input/output types for study operations.
package study

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// MaxExcludeIDsCount caps the size of the client-supplied resume set
// (questions already answered in the in-progress browser session) so a
// malicious or buggy client cannot send an unbounded list.
const MaxExcludeIDsCount = 1000

// MaxExcludeIDLength caps each entry in the resume set. Mirrors
// MaxChoiceIDLength so the two question-ID inputs hit the same upper bound,
// keeping per-request memory bounded even if the count cap is reached.
const MaxExcludeIDLength = 100

// GetStudyQuestionsInput is the validated input for getting study questions.
//
// Practice is an off-the-record mode: when true, the usecase ignores the
// per-question NextDueAt schedule and returns every active question. Callers
// using this mode are expected to skip the "record answer" endpoint so the
// user's spaced-repetition records and counters stay untouched.
//
// ExcludeIDs lets the client resume an interrupted study session by skipping
// questions it has already answered locally. The server is stateless; the
// browser tracks the in-progress set and discards it at the next local 03:00
// boundary.
type GetStudyQuestionsInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Limit          int    `validate:"gte=1,lte=100"`
	Practice       bool
	ExcludeIDs     []string
}

// GetStudyQuestionsInputParams holds the named parameters for NewGetStudyQuestionsInput.
type GetStudyQuestionsInputParams struct {
	OperatorID     string
	OrganizationID string
	WorkbookID     string
	Limit          int
	Practice       bool
	ExcludeIDs     []string
}

// NewGetStudyQuestionsInput creates a validated GetStudyQuestionsInput.
func NewGetStudyQuestionsInput(p GetStudyQuestionsInputParams) (*GetStudyQuestionsInput, error) {
	if len(p.ExcludeIDs) > MaxExcludeIDsCount {
		return nil, fmt.Errorf("excludeIds count exceeds limit (max %d, got %d): %w", MaxExcludeIDsCount, len(p.ExcludeIDs), domain.ErrInvalidArgument)
	}
	for i, id := range p.ExcludeIDs {
		if id == "" {
			return nil, fmt.Errorf("excludeIds[%d] is empty: %w", i, domain.ErrInvalidArgument)
		}
		if len(id) > MaxExcludeIDLength {
			return nil, fmt.Errorf("excludeIds[%d] exceeds length limit (max %d, got %d): %w", i, MaxExcludeIDLength, len(id), domain.ErrInvalidArgument)
		}
	}
	m := &GetStudyQuestionsInput{
		OperatorID:     p.OperatorID,
		OrganizationID: p.OrganizationID,
		WorkbookID:     p.WorkbookID,
		Limit:          p.Limit,
		Practice:       p.Practice,
		ExcludeIDs:     p.ExcludeIDs,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get study questions input: %w", err)
	}
	return m, nil
}

// GetStudySummaryInput is the validated input for getting study summary counts.
//
// Practice mirrors GetStudyQuestionsInput.Practice: when true the summary is
// computed without applying the per-question NextDueAt schedule, so callers
// rendering a "practice mode" picker see the unrestricted available pool.
type GetStudySummaryInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
	Practice       bool
}

// GetStudySummaryInputParams holds the named parameters for NewGetStudySummaryInput.
type GetStudySummaryInputParams struct {
	OperatorID     string
	OrganizationID string
	WorkbookID     string
	Practice       bool
}

// NewGetStudySummaryInput creates a validated GetStudySummaryInput.
func NewGetStudySummaryInput(p GetStudySummaryInputParams) (*GetStudySummaryInput, error) {
	m := &GetStudySummaryInput{
		OperatorID:     p.OperatorID,
		OrganizationID: p.OrganizationID,
		WorkbookID:     p.WorkbookID,
		Practice:       p.Practice,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get study summary input: %w", err)
	}
	return m, nil
}

// GetStudySummaryOutput is the output for getting study summary counts. The
// ratio fields advertise the server-side review/new mix so clients can render
// it in the picker without hard-coding the same constants.
type GetStudySummaryOutput struct {
	NewCount               int
	ReviewCount            int
	TotalDue               int
	ReviewRatioNumerator   int
	ReviewRatioDenominator int
}

// QuestionItem represents a question returned for study.
type QuestionItem struct {
	QuestionID   string
	QuestionType string
	Content      string
	Tags         []string
	OrderIndex   int
	// Audio carries playable audio metadata when the audio batch has finished.
	// Nil when no audio is ready (the typical state for a freshly created
	// word_fill question).
	Audio *QuestionItemAudio
}

// QuestionItemAudio mirrors the same shape as questionservice.AudioOutput but
// avoids cross-package coupling: study items are independent of the question
// CRUD response shape.
type QuestionItemAudio struct {
	Source *QuestionItemAudioRef
	Target *QuestionItemAudioRef
}

// QuestionItemAudioRef carries the storage-relative path and metadata of a
// single audio file. The handler turns Path into a public URL.
type QuestionItemAudioRef struct {
	Path        string
	DurationSec float64
}

// GetStudyQuestionsOutput is the output for getting study questions.
type GetStudyQuestionsOutput struct {
	Questions   []QuestionItem
	TotalDue    int
	NewCount    int
	ReviewCount int
}

// Limits on the multiple_choice answer payload. These mirror the OpenAPI
// schema (maxItems / maxLength on selectedChoiceIds) and are enforced here so
// the contract holds even when callers bypass the spec.
const (
	MaxSelectedChoiceIDsCount = 40
	MaxChoiceIDLength         = 100
)

// RecordAnswerInput is the validated input for recording an answer.
// Exactly one of Correct or SelectedChoiceIDs must be set, matched to the
// question's type. The usecase enforces the per-type rule once the question
// is loaded (the handler also rejects "neither set" / "both set" earlier).
//
// LocalDateKey and Timezone are the client-supplied user-local "today" used
// to bucket the answer into the dashboard's daily contribution graph. They
// are optional: when empty, the daily-stats increment is skipped (the answer
// itself is still recorded). The handler reads these from the X-Local-Date
// and X-Local-Timezone headers respectively, which keeps server clocks and
// user TZ preferences independent of the SRS write path.
type RecordAnswerInput struct {
	OperatorID        string `validate:"required"`
	OrganizationID    string `validate:"required"`
	WorkbookID        string `validate:"required"`
	QuestionID        string `validate:"required"`
	Correct           *bool
	SelectedChoiceIDs *[]string
	LocalDateKey      string
	Timezone          string
}

// RecordAnswerInputForWordFillParams holds the named parameters for NewRecordAnswerInputForWordFill.
type RecordAnswerInputForWordFillParams struct {
	OperatorID     string
	OrganizationID string
	WorkbookID     string
	QuestionID     string
	Correct        bool
	LocalDateKey   string
	Timezone       string
}

// NewRecordAnswerInputForWordFill creates a validated RecordAnswerInput for word_fill questions.
func NewRecordAnswerInputForWordFill(p RecordAnswerInputForWordFillParams) (*RecordAnswerInput, error) {
	m := &RecordAnswerInput{
		OperatorID:        p.OperatorID,
		OrganizationID:    p.OrganizationID,
		WorkbookID:        p.WorkbookID,
		QuestionID:        p.QuestionID,
		Correct:           &p.Correct,
		SelectedChoiceIDs: nil,
		LocalDateKey:      p.LocalDateKey,
		Timezone:          p.Timezone,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate record answer input: %w", err)
	}
	return m, nil
}

// RecordAnswerInputForMultipleChoiceParams holds the named parameters for NewRecordAnswerInputForMultipleChoice.
type RecordAnswerInputForMultipleChoiceParams struct {
	OperatorID        string
	OrganizationID    string
	WorkbookID        string
	QuestionID        string
	SelectedChoiceIDs []string
	LocalDateKey      string
	Timezone          string
}

// NewRecordAnswerInputForMultipleChoice creates a validated RecordAnswerInput for multiple_choice questions.
func NewRecordAnswerInputForMultipleChoice(p RecordAnswerInputForMultipleChoiceParams) (*RecordAnswerInput, error) {
	if len(p.SelectedChoiceIDs) > MaxSelectedChoiceIDsCount {
		return nil, fmt.Errorf("selectedChoiceIds count exceeds limit (max %d, got %d): %w", MaxSelectedChoiceIDsCount, len(p.SelectedChoiceIDs), domain.ErrInvalidArgument)
	}
	for i, id := range p.SelectedChoiceIDs {
		if len(id) > MaxChoiceIDLength {
			return nil, fmt.Errorf("selectedChoiceIds[%d] exceeds length limit (max %d, got %d): %w", i, MaxChoiceIDLength, len(id), domain.ErrInvalidArgument)
		}
	}
	ids := p.SelectedChoiceIDs
	m := &RecordAnswerInput{
		OperatorID:        p.OperatorID,
		OrganizationID:    p.OrganizationID,
		WorkbookID:        p.WorkbookID,
		QuestionID:        p.QuestionID,
		Correct:           nil,
		SelectedChoiceIDs: &ids,
		LocalDateKey:      p.LocalDateKey,
		Timezone:          p.Timezone,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate record answer input: %w", err)
	}
	return m, nil
}

// RecordAnswerOutput is the output for recording an answer.
type RecordAnswerOutput struct {
	NextDueAt          time.Time
	ConsecutiveCorrect int
	TotalCorrect       int
	TotalIncorrect     int
}

// DeleteStudyHistoryInput is the validated input for clearing a workbook's
// study history. The operation only affects the operator's own records.
type DeleteStudyHistoryInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// DeleteStudyHistoryInputParams holds the named parameters for NewDeleteStudyHistoryInput.
type DeleteStudyHistoryInputParams struct {
	OperatorID     string
	OrganizationID string
	WorkbookID     string
}

// NewDeleteStudyHistoryInput creates a validated DeleteStudyHistoryInput.
func NewDeleteStudyHistoryInput(p DeleteStudyHistoryInputParams) (*DeleteStudyHistoryInput, error) {
	m := &DeleteStudyHistoryInput{
		OperatorID:     p.OperatorID,
		OrganizationID: p.OrganizationID,
		WorkbookID:     p.WorkbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate delete study history input: %w", err)
	}
	return m, nil
}

// ListStudyRecordsInput is the validated input for listing a workbook's
// study records belonging to the operator.
type ListStudyRecordsInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	WorkbookID     string `validate:"required"`
}

// ListStudyRecordsInputParams holds the named parameters for NewListStudyRecordsInput.
type ListStudyRecordsInputParams struct {
	OperatorID     string
	OrganizationID string
	WorkbookID     string
}

// NewListStudyRecordsInput creates a validated ListStudyRecordsInput.
func NewListStudyRecordsInput(p ListStudyRecordsInputParams) (*ListStudyRecordsInput, error) {
	m := &ListStudyRecordsInput{
		OperatorID:     p.OperatorID,
		OrganizationID: p.OrganizationID,
		WorkbookID:     p.WorkbookID,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate list study records input: %w", err)
	}
	return m, nil
}

// RecordItem mirrors a single study record at the service boundary,
// decoupling controllers from the domain aggregate.
type RecordItem struct {
	WorkbookID         string
	QuestionID         string
	ConsecutiveCorrect int
	LastAnsweredAt     time.Time
	NextDueAt          time.Time
	TotalCorrect       int
	TotalIncorrect     int
}

// ListStudyRecordsOutput holds the records returned for a workbook.
type ListStudyRecordsOutput struct {
	Records []RecordItem
}

// MinDashboardDays / MaxDashboardDays bound the requested contribution-graph
// window. 7 supports a one-week mini-view; 730 supports the GitHub-style
// "two-year" maximum, which is more than enough for the canonical 365-day
// graph without leaving headroom for accidentally-unbounded ranges.
const (
	MinDashboardDays = 7
	MaxDashboardDays = 730
)

// GetDashboardInput is the validated input for the user-scoped study
// dashboard. TodayDateKey is the client's "today" in YYYY-MM-DD form
// (the frontend computes it in the user's local timezone before sending
// X-Local-Date) so server clocks stay decoupled from the user's local
// calendar; the usecase derives the [today - days + 1, today] window
// purely from the date string and does not need the timezone here.
type GetDashboardInput struct {
	OperatorID     string `validate:"required"`
	OrganizationID string `validate:"required"`
	Days           int    `validate:"gte=7,lte=730"`
	TodayDateKey   string `validate:"required"`
}

// GetDashboardInputParams holds the named parameters for NewGetDashboardInput.
type GetDashboardInputParams struct {
	OperatorID     string
	OrganizationID string
	Days           int
	TodayDateKey   string
}

// NewGetDashboardInput creates a validated GetDashboardInput.
func NewGetDashboardInput(p GetDashboardInputParams) (*GetDashboardInput, error) {
	m := &GetDashboardInput{
		OperatorID:     p.OperatorID,
		OrganizationID: p.OrganizationID,
		Days:           p.Days,
		TodayDateKey:   p.TodayDateKey,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate get dashboard input: %w", err)
	}
	return m, nil
}

// DashboardDailyItem mirrors a single daily bucket on the service boundary.
type DashboardDailyItem struct {
	Date          string
	AnsweredCount int
	CorrectCount  int
}

// GetDashboardOutput holds the data needed to render the dashboard: the
// per-day contribution buckets within the requested window, the rolling
// streak counters (calendar-day based, requiring at least one answer per
// day), today's running total, and the all-window aggregates.
type GetDashboardOutput struct {
	From          string
	To            string
	Days          []DashboardDailyItem
	CurrentStreak int
	LongestStreak int
	TodayCount    int
	TodayCorrect  int
	ActiveDays    int
	TotalAnswered int
	TotalCorrect  int
}
