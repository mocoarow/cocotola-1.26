package seed

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

const (
	// seedKeyMarker is appended to the workbook description as a hidden marker
	// so the seeder can identify previously-seeded workbooks even if their
	// human-readable Title is later changed in the seed file.
	seedKeyMarker = "[seedKey:"
	seedKeyEnd    = "]"
	// questionTagPrefix is added to each question's Tags slice for the same
	// idempotency purpose. Format: seed:<workbookSeedKey>:<questionSeedKey>.
	questionTagPrefix = "seed:"
)

// WorkbookAPIClient is the subset of QuestionAPIClient required by WorkbookSeeder.
// Defining it here lets the seeder be tested without spinning up an HTTP server.
type WorkbookAPIClient interface {
	ListWorkbooks(ctx context.Context, organizationID, spaceID string) ([]WorkbookListItem, error)
	CreateWorkbook(ctx context.Context, organizationID string, body CreateWorkbookRequest) (string, error)
	ListQuestions(ctx context.Context, organizationID, workbookID string) ([]QuestionListItem, error)
	AddQuestion(ctx context.Context, organizationID, workbookID string, body AddQuestionRequest) error
}

// WorkbookSeeder applies a list of PublicWorkbookSeed against the question
// service in an idempotent manner.
type WorkbookSeeder struct {
	client WorkbookAPIClient
	seeds  []PublicWorkbookSeed
	logger *slog.Logger
}

// NewWorkbookSeeder constructs a seeder using the given client and seeds.
func NewWorkbookSeeder(client WorkbookAPIClient, seeds []PublicWorkbookSeed) *WorkbookSeeder {
	return &WorkbookSeeder{
		client: client,
		seeds:  seeds,
		logger: slog.Default().With(slog.String("component", "public-workbook-seeder")),
	}
}

// SeedPublicWorkbooks creates each seed's workbook (if absent) and then each
// nested question (if absent). Existing entities are detected via seedKey markers.
func (s *WorkbookSeeder) SeedPublicWorkbooks(ctx context.Context, organizationID, publicSpaceID string) error {
	existing, err := s.client.ListWorkbooks(ctx, organizationID, publicSpaceID)
	if err != nil {
		return fmt.Errorf("list existing workbooks: %w", err)
	}

	bySeedKey := indexWorkbooksBySeedKey(existing)

	for _, sd := range s.seeds {
		workbookID, err := s.ensureWorkbook(ctx, organizationID, publicSpaceID, sd, bySeedKey)
		if err != nil {
			return fmt.Errorf("ensure workbook %q: %w", sd.SeedKey, err)
		}

		if err := s.ensureQuestions(ctx, organizationID, workbookID, sd); err != nil {
			return fmt.Errorf("ensure questions for workbook %q: %w", sd.SeedKey, err)
		}
	}

	return nil
}

func (s *WorkbookSeeder) ensureWorkbook(ctx context.Context, organizationID, publicSpaceID string, sd PublicWorkbookSeed, existing map[string]string) (string, error) {
	if id, ok := existing[sd.SeedKey]; ok {
		s.logger.InfoContext(ctx, "workbook already seeded",
			slog.String("seed_key", sd.SeedKey),
			slog.String("workbook_id", id),
		)
		return id, nil
	}

	body := CreateWorkbookRequest{
		SpaceID:     publicSpaceID,
		Title:       sd.Title,
		Description: encodeDescription(sd.Description, sd.SeedKey),
		Visibility:  "public",
	}
	workbookID, err := s.client.CreateWorkbook(ctx, organizationID, body)
	if err != nil {
		return "", fmt.Errorf("create workbook %q: %w", sd.SeedKey, err)
	}

	s.logger.InfoContext(ctx, "workbook created",
		slog.String("seed_key", sd.SeedKey),
		slog.String("workbook_id", workbookID),
	)
	return workbookID, nil
}

func (s *WorkbookSeeder) ensureQuestions(ctx context.Context, organizationID, workbookID string, sd PublicWorkbookSeed) error {
	if len(sd.Questions) == 0 {
		return nil
	}

	existing, err := s.client.ListQuestions(ctx, organizationID, workbookID)
	if err != nil {
		return fmt.Errorf("list existing questions: %w", err)
	}
	seenTags := indexQuestionTags(existing)

	for _, q := range sd.Questions {
		tag := questionTag(sd.SeedKey, q.SeedKey)
		if seenTags[tag] {
			s.logger.InfoContext(ctx, "question already seeded",
				slog.String("workbook_seed_key", sd.SeedKey),
				slog.String("question_seed_key", q.SeedKey),
			)

			continue
		}

		body := AddQuestionRequest{
			QuestionType: q.QuestionType,
			Content:      q.Content,
			Tags:         append([]string{tag}, q.Tags...),
			OrderIndex:   q.OrderIndex,
		}
		if err := s.client.AddQuestion(ctx, organizationID, workbookID, body); err != nil {
			return fmt.Errorf("add question %q: %w", q.SeedKey, err)
		}

		s.logger.InfoContext(ctx, "question created",
			slog.String("workbook_seed_key", sd.SeedKey),
			slog.String("question_seed_key", q.SeedKey),
		)
	}

	return nil
}

// indexWorkbooksBySeedKey extracts the seedKey marker from each workbook's
// description and returns a map seedKey -> workbookID for fast lookup.
func indexWorkbooksBySeedKey(items []WorkbookListItem) map[string]string {
	result := make(map[string]string, len(items))
	for _, w := range items {
		if key, ok := decodeSeedKey(w.Description); ok {
			result[key] = w.WorkbookID
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

// encodeDescription appends the seedKey marker to the description so it can be
// recovered by decodeSeedKey on subsequent runs.
func encodeDescription(description, seedKey string) string {
	marker := seedKeyMarker + seedKey + seedKeyEnd
	if description == "" {
		return marker
	}
	return description + " " + marker
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
