package seed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/i18n"
)

const (
	seedKeyMarker = "[seedKey:"
	seedKeyEnd    = "]"
	// questionTagPrefix forms seed-<workbookSeedKey>:<questionSeedKey>; a hyphen,
	// not a colon, because cocotola-question tags allow only one colon.
	questionTagPrefix = "seed-"
)

// ErrQuestionRejected is returned by WorkbookAPIClient.AddQuestion when the
// question service rejects the question as invalid.
var ErrQuestionRejected = errors.New("question rejected")

// ErrSeedKeyNotStored is returned when the question service does not store the seed key.
var ErrSeedKeyNotStored = errors.New("question service did not store the seed key")

// WorkbookAPIClient is the question service API required by WorkbookSeeder.
type WorkbookAPIClient interface {
	ListWorkbooks(ctx context.Context, organizationID, spaceID string) ([]WorkbookListItem, error)
	CreateWorkbook(ctx context.Context, organizationID string, input CreateWorkbookInput) (CreateWorkbookOutput, error)
	UpdateWorkbook(ctx context.Context, organizationID, workbookID string, input UpdateWorkbookInput) (seedKey string, err error)
	ListQuestions(ctx context.Context, organizationID, workbookID string) ([]QuestionListItem, error)
	AddQuestion(ctx context.Context, organizationID, workbookID string, input AddQuestionInput) error
}

// WorkbookListItem is a workbook already present in the space.
type WorkbookListItem struct {
	WorkbookID  string
	Title       string
	Description string
	Visibility  string
	Language    string
	SeedKey     string
}

// CreateWorkbookInput describes a workbook to create.
type CreateWorkbookInput struct {
	SpaceID     string
	Title       string
	Description string
	Language    string
	SeedKey     string
}

// CreateWorkbookOutput is a created workbook.
type CreateWorkbookOutput struct {
	WorkbookID string
	SeedKey    string
}

// UpdateWorkbookInput describes a workbook update.
type UpdateWorkbookInput struct {
	Title       string
	Description string
	Visibility  string
	Language    string
	SeedKey     string
}

// QuestionListItem is a question already present in the workbook.
type QuestionListItem struct {
	QuestionID string
	Tags       []string
}

// AddQuestionInput describes a question to add.
type AddQuestionInput struct {
	QuestionType string
	Content      string
	Tags         []string
	OrderIndex   int32
}

// WorkbookPolicyEnsurer idempotently grants the system owner per-workbook RBAC
// policies. It runs for every workbook, existing ones included, so workbooks
// left without policies by a failed run are repaired instead of staying 403.
type WorkbookPolicyEnsurer interface {
	EnsureSystemOwnerWorkbookPolicies(ctx context.Context, organizationID, workbookID string) error
}

// WorkbookSeeder applies a list of PublicWorkbookSeed against the question
// service in an idempotent manner.
type WorkbookSeeder struct {
	client        WorkbookAPIClient
	policyEnsurer WorkbookPolicyEnsurer
	logger        *slog.Logger
}

// NewWorkbookSeeder constructs a seeder using the given client, policy ensurer and logger.
func NewWorkbookSeeder(client WorkbookAPIClient, policyEnsurer WorkbookPolicyEnsurer, logger *slog.Logger) *WorkbookSeeder {
	return &WorkbookSeeder{
		client:        client,
		policyEnsurer: policyEnsurer,
		logger:        logger,
	}
}

// PublicWorkbookSeeder seeds a fixed list of public workbooks.
type PublicWorkbookSeeder struct {
	seeder *WorkbookSeeder
	seeds  []PublicWorkbookSeed
}

// NewPublicWorkbookSeeder constructs a PublicWorkbookSeeder that applies seeds via seeder.
func NewPublicWorkbookSeeder(seeder *WorkbookSeeder, seeds []PublicWorkbookSeed) *PublicWorkbookSeeder {
	return &PublicWorkbookSeeder{
		seeder: seeder,
		seeds:  seeds,
	}
}

// WorkbookOutput is the result of seeding one workbook.
// MaxExistingOrderIndex is the largest OrderIndex among the seed's questions
// that already existed before this run.
type WorkbookOutput struct {
	SeedKey               string
	AddedOrderIndexes     []int32
	RejectedOrderIndexes  []int32
	SkippedExisting       int
	MaxExistingOrderIndex int32
}

// SeedPublicWorkbooks seeds the workbooks given to NewPublicWorkbookSeeder and
// returns ErrInvalidSeed if the question service rejected any of their questions.
func (s *PublicWorkbookSeeder) SeedPublicWorkbooks(ctx context.Context, organizationID, publicSpaceID string) error {
	outputs, err := s.seeder.SeedWorkbooks(ctx, organizationID, publicSpaceID, s.seeds)
	if err != nil {
		return err
	}

	var errs []error
	for _, out := range outputs {
		if len(out.RejectedOrderIndexes) > 0 {
			errs = append(errs, fmt.Errorf("workbook %q: question service rejected questions at order indexes %v: %w",
				out.SeedKey, out.RejectedOrderIndexes, ErrInvalidSeed))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("embedded seed questions rejected: %w", errors.Join(errs...))
	}
	return nil
}

// SeedWorkbooks creates each seed's workbook (if absent) and then each nested
// question (if absent). Existing workbooks are detected by their seed key.
func (s *WorkbookSeeder) SeedWorkbooks(ctx context.Context, organizationID, publicSpaceID string, seeds []PublicWorkbookSeed) ([]WorkbookOutput, error) {
	existing, err := s.client.ListWorkbooks(ctx, organizationID, publicSpaceID)
	if err != nil {
		return nil, fmt.Errorf("list existing workbooks: %w", err)
	}

	bySeedKey := indexWorkbooksBySeedKey(existing)

	outputs := make([]WorkbookOutput, 0, len(seeds))
	for _, sd := range seeds {
		existingWorkbook, found := bySeedKey[sd.SeedKey]
		workbookID, err := s.ensureWorkbook(ctx, organizationID, publicSpaceID, sd, existingWorkbook, found)
		if err != nil {
			return nil, fmt.Errorf("ensure workbook %q: %w", sd.SeedKey, err)
		}

		// Grant the per-workbook policies before adding questions. This both
		// covers freshly created workbooks and repairs ones that an earlier run
		// persisted without policies, so AddQuestion authorization succeeds.
		if err := s.policyEnsurer.EnsureSystemOwnerWorkbookPolicies(ctx, organizationID, workbookID); err != nil {
			return nil, fmt.Errorf("ensure workbook policies %q: %w", sd.SeedKey, err)
		}

		if found {
			if err := s.migrateLegacySeedKey(ctx, organizationID, sd.SeedKey, existingWorkbook); err != nil {
				return nil, fmt.Errorf("migrate seed key of workbook %q: %w", sd.SeedKey, err)
			}
		}

		output, err := s.ensureQuestions(ctx, organizationID, workbookID, sd)
		if err != nil {
			return nil, fmt.Errorf("ensure questions for workbook %q: %w", sd.SeedKey, err)
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *WorkbookSeeder) ensureWorkbook(ctx context.Context, organizationID, publicSpaceID string, sd PublicWorkbookSeed, existing WorkbookListItem, found bool) (string, error) {
	if found {
		s.logger.InfoContext(ctx, "workbook already seeded",
			slog.String("seed_key", sd.SeedKey),
			slog.String("workbook_id", existing.WorkbookID),
		)
		return existing.WorkbookID, nil
	}

	// Catch malformed `language` early — without this, the cocotola-question
	// API rejects the request with a generic 400 that is hard to map back
	// to the offending seed entry. Reject anything outside the canonical
	// lowercase ISO 639-1 form (e.g. "JA", "jpn", "ja-JP").
	if !i18n.IsValidISO6391(sd.Language) {
		return "", fmt.Errorf("seed %q: language %q must be a lowercase ISO 639-1 code: %w", sd.SeedKey, sd.Language, ErrInvalidSeed)
	}

	input := CreateWorkbookInput{
		SpaceID:     publicSpaceID,
		Title:       sd.Title,
		Description: sd.Description,
		Language:    sd.Language,
		SeedKey:     sd.SeedKey,
	}
	created, err := s.client.CreateWorkbook(ctx, organizationID, input)
	if err != nil {
		return "", fmt.Errorf("create workbook %q: %w", sd.SeedKey, err)
	}
	workbookID := created.WorkbookID
	if created.SeedKey != sd.SeedKey {
		return "", fmt.Errorf("create workbook %q: workbook %s was created with seed key %q: %w", sd.SeedKey, workbookID, created.SeedKey, ErrSeedKeyNotStored)
	}

	s.logger.InfoContext(ctx, "workbook created",
		slog.String("seed_key", sd.SeedKey),
		slog.String("workbook_id", workbookID),
	)
	return workbookID, nil
}

func (s *WorkbookSeeder) migrateLegacySeedKey(ctx context.Context, organizationID, seedKey string, wb WorkbookListItem) error {
	if wb.SeedKey == "" {
		input := UpdateWorkbookInput{Title: wb.Title, Description: wb.Description, Visibility: wb.Visibility, Language: wb.Language, SeedKey: seedKey}
		stored, err := s.client.UpdateWorkbook(ctx, organizationID, wb.WorkbookID, input)
		if err != nil {
			return fmt.Errorf("assign seed key to workbook %s: %w", wb.WorkbookID, err)
		}
		if stored != seedKey {
			return fmt.Errorf("assign seed key to workbook %s: got seed key %q: %w", wb.WorkbookID, stored, ErrSeedKeyNotStored)
		}
		s.logMigrated(ctx, seedKey, wb.WorkbookID, "assign_seed_key")
	}

	description, found := removeSeedKeyMarker(wb.Description, seedKey)
	if !found {
		return nil
	}
	input := UpdateWorkbookInput{Title: wb.Title, Description: description, Visibility: wb.Visibility, Language: wb.Language, SeedKey: seedKey}
	stored, err := s.client.UpdateWorkbook(ctx, organizationID, wb.WorkbookID, input)
	if err != nil {
		return fmt.Errorf("remove seed key marker from workbook %s: %w", wb.WorkbookID, err)
	}
	if stored != seedKey {
		return s.restoreSeedKeyMarker(ctx, organizationID, seedKey, wb, stored)
	}
	s.logMigrated(ctx, seedKey, wb.WorkbookID, "remove_marker")
	return nil
}

func (s *WorkbookSeeder) restoreSeedKeyMarker(ctx context.Context, organizationID, seedKey string, wb WorkbookListItem, stored string) error {
	notStored := fmt.Errorf("remove seed key marker from workbook %s: got seed key %q: %w", wb.WorkbookID, stored, ErrSeedKeyNotStored)
	input := UpdateWorkbookInput{Title: wb.Title, Description: wb.Description, Visibility: wb.Visibility, Language: wb.Language, SeedKey: seedKey}
	if _, err := s.client.UpdateWorkbook(ctx, organizationID, wb.WorkbookID, input); err != nil {
		return fmt.Errorf("restore seed key marker of workbook %s: %w", wb.WorkbookID, errors.Join(notStored, err))
	}
	return notStored
}

func (s *WorkbookSeeder) logMigrated(ctx context.Context, seedKey, workbookID, step string) {
	s.logger.InfoContext(ctx, "workbook seed key migrated",
		slog.String("seed_key", seedKey),
		slog.String("workbook_id", workbookID),
		slog.String("step", step),
	)
}

func (s *WorkbookSeeder) ensureQuestions(ctx context.Context, organizationID, workbookID string, sd PublicWorkbookSeed) (WorkbookOutput, error) {
	output := WorkbookOutput{SeedKey: sd.SeedKey, AddedOrderIndexes: nil, RejectedOrderIndexes: nil, SkippedExisting: 0, MaxExistingOrderIndex: 0}
	if len(sd.Questions) == 0 {
		return output, nil
	}

	existing, err := s.client.ListQuestions(ctx, organizationID, workbookID)
	if err != nil {
		return WorkbookOutput{}, fmt.Errorf("list existing questions: %w", err)
	}
	seenTags := indexQuestionTags(existing)

	for _, q := range sd.Questions {
		tag := questionTag(sd.SeedKey, q.SeedKey)
		if seenTags[tag] {
			output.SkippedExisting++
			output.MaxExistingOrderIndex = max(output.MaxExistingOrderIndex, q.OrderIndex)
			s.logger.DebugContext(ctx, "question already seeded",
				slog.String("workbook_seed_key", sd.SeedKey),
				slog.String("question_seed_key", q.SeedKey),
			)

			continue
		}

		wasRejected, err := s.addQuestion(ctx, organizationID, workbookID, sd.SeedKey, tag, q)
		if err != nil {
			return WorkbookOutput{}, err
		}
		if wasRejected {
			output.RejectedOrderIndexes = append(output.RejectedOrderIndexes, q.OrderIndex)
		} else {
			output.AddedOrderIndexes = append(output.AddedOrderIndexes, q.OrderIndex)
		}
	}

	s.logger.DebugContext(ctx, "workbook questions seeded",
		slog.String("workbook_seed_key", sd.SeedKey),
		slog.Int("added", len(output.AddedOrderIndexes)),
		slog.Int("skipped_existing", output.SkippedExisting),
		slog.Int("rejected", len(output.RejectedOrderIndexes)),
	)
	return output, nil
}

// addQuestion adds one question. A rejected question will be rejected on every
// run, so it is reported as rejected instead of stopping the remaining questions.
func (s *WorkbookSeeder) addQuestion(ctx context.Context, organizationID, workbookID, workbookSeedKey, tag string, q QuestionSeed) (bool, error) {
	input := AddQuestionInput{
		QuestionType: q.QuestionType,
		Content:      q.Content,
		Tags:         append([]string{tag}, q.Tags...),
		OrderIndex:   q.OrderIndex,
	}
	err := s.client.AddQuestion(ctx, organizationID, workbookID, input)
	if errors.Is(err, ErrQuestionRejected) {
		s.logger.WarnContext(ctx, "question rejected",
			slog.String("workbook_seed_key", workbookSeedKey),
			slog.String("question_seed_key", q.SeedKey),
		)
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("add question %q: %w", q.SeedKey, err)
	}

	s.logger.DebugContext(ctx, "question created",
		slog.String("workbook_seed_key", workbookSeedKey),
		slog.String("question_seed_key", q.SeedKey),
	)
	return false, nil
}

func indexWorkbooksBySeedKey(items []WorkbookListItem) map[string]WorkbookListItem {
	result := make(map[string]WorkbookListItem, len(items))
	for _, w := range items {
		if _, taken := result[w.SeedKey]; w.SeedKey != "" && !taken {
			result[w.SeedKey] = w
		}
	}
	for _, w := range items {
		if w.SeedKey != "" {
			continue
		}
		key, ok := decodeSeedKey(w.Description)
		if !ok {
			continue
		}
		if prev, taken := result[key]; !taken || prev.SeedKey == "" {
			result[key] = w
		}
	}
	return result
}

func indexQuestionTags(items []QuestionListItem) map[string]bool {
	result := make(map[string]bool)
	for _, q := range items {
		for _, t := range q.Tags {
			if strings.HasPrefix(t, questionTagPrefix) {
				result[t] = true
			}
		}
	}
	return result
}

func removeSeedKeyMarker(description, seedKey string) (string, bool) {
	marker := seedKeyMarker + seedKey + seedKeyEnd
	if !strings.Contains(description, marker) {
		return description, false
	}
	if strings.Contains(description, " "+marker) {
		return strings.Replace(description, " "+marker, "", 1), true
	}
	return strings.Replace(description, marker, "", 1), true
}

// decodeSeedKey extracts the seedKey marker from a description string. Returns
// ("", false) when the marker is absent or malformed.
func decodeSeedKey(description string) (string, bool) {
	_, after, found := strings.Cut(description, seedKeyMarker)
	if !found {
		return "", false
	}
	key, _, found := strings.Cut(after, seedKeyEnd)
	if !found || key == "" {
		return "", false
	}
	return key, true
}

func questionTag(workbookSeedKey, questionSeedKey string) string {
	return questionTagPrefix + workbookSeedKey + ":" + questionSeedKey
}
