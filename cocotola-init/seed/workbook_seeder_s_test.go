//go:build small

package seed_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

// questionTagPattern mirrors the validation regex used by cocotola-question's
// domain layer. It is duplicated here on purpose so the seed package does not
// take a hard dependency on the question domain just for this regression test.
var questionTagPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+$`)

const (
	testOrgID         = "org-1"
	testPublicSpaceID = "public-space-1"
)

var errQuestionRejected = fmt.Errorf("add question: %w", seed.ErrQuestionRejected)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func sampleSeeds() []seed.PublicWorkbookSeed {
	return []seed.PublicWorkbookSeed{
		{
			SeedKey:     "vocab-v1",
			Title:       "Vocabulary",
			Description: "vocab desc",
			Language:    "ja",
			Questions: []seed.QuestionSeed{
				{SeedKey: "q1", QuestionType: "word_fill", Content: "C1", OrderIndex: 0},
				{SeedKey: "q2", QuestionType: "word_fill", Content: "C2", OrderIndex: 1},
			},
		},
		{
			SeedKey:     "grammar-v1",
			Title:       "Grammar",
			Description: "grammar desc",
			Language:    "ja",
		},
	}
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldCreateAllWorkbooksAndQuestions_onFirstRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(input seed.CreateWorkbookInput) bool {
		return input.Title == "Vocabulary" && strings.Contains(input.Description, "[seedKey:vocab-v1]")
	})).Return("wb-Vocabulary", nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(input seed.CreateWorkbookInput) bool {
		return input.Title == "Grammar" && strings.Contains(input.Description, "[seedKey:grammar-v1]")
	})).Return("wb-Grammar", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.MatchedBy(func(input seed.AddQuestionInput) bool {
		return input.Content == "C1"
	})).Return(nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.MatchedBy(func(input seed.AddQuestionInput) bool {
		return input.Content == "C2"
	})).Return(nil)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Vocabulary").Return(nil)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Grammar").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldBeIdempotent_onSecondRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: simulate second run where workbooks and questions already exist
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{
			{WorkbookID: "wb-Vocabulary", Title: "Vocabulary", Description: "vocab desc [seedKey:vocab-v1]"},
			{WorkbookID: "wb-Grammar", Title: "Grammar", Description: "grammar desc [seedKey:grammar-v1]"},
		}, nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return([]seed.QuestionListItem{
			{QuestionID: "q-1", Tags: []string{"seed-vocab-v1:q1"}},
			{QuestionID: "q-2", Tags: []string{"seed-vocab-v1:q2"}},
		}, nil)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Vocabulary").Return(nil)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Grammar").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: no CreateWorkbook or AddQuestion calls
	require.NoError(t, err)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldDetectExistingWorkbook_byDescriptionSeedKey_evenWhenTitleChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a workbook already exists with the right seedKey but a different title
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{{
			WorkbookID:  "wb-existing",
			Title:       "Old Vocabulary Title",
			Description: "any free text [seedKey:vocab-v1]",
		}}, nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-existing").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(input seed.AddQuestionInput) bool {
		return input.Content == "C1"
	})).Return(nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(input seed.AddQuestionInput) bool {
		return input.Content == "C2"
	})).Return(nil)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-existing").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: the workbook is reused (not recreated) and its questions are added
	require.NoError(t, err)
}

// csvSeeds returns a workbook seed whose questions carry non-zero OrderIndex
// values, like seeds converted from a CSV.
func csvSeeds() []seed.PublicWorkbookSeed {
	return []seed.PublicWorkbookSeed{{
		SeedKey:  "vocab-v1",
		Title:    "Vocabulary",
		Language: "ja",
		Questions: []seed.QuestionSeed{
			{SeedKey: "q1", QuestionType: "word_fill", Content: "C1", OrderIndex: 5},
			{SeedKey: "q2", QuestionType: "word_fill", Content: "C2", OrderIndex: 7},
			{SeedKey: "q3", QuestionType: "word_fill", Content: "C3", OrderIndex: 9},
		},
	}}
}

func Test_WorkbookSeeder_SeedWorkbooks_shouldReportPositionsAndMaxExistingOrderIndex_whenSomeQuestionsExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: q1 and q2 were imported before
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{{WorkbookID: "wb-existing", Description: "[seedKey:vocab-v1]"}}, nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-existing").
		Return([]seed.QuestionListItem{
			{QuestionID: "q-1", Tags: []string{"seed-vocab-v1:q1"}},
			{QuestionID: "q-2", Tags: []string{"seed-vocab-v1:q2"}},
		}, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C3")).Return(nil)
	seeder := seed.NewWorkbookSeeder(client, existingVocabularyEnsurer(ctx, t), discardLogger())

	// when
	got, err := seeder.SeedWorkbooks(ctx, testOrgID, testPublicSpaceID, csvSeeds())

	// then
	require.NoError(t, err)
	assert.Equal(t, []seed.WorkbookOutput{
		{SeedKey: "vocab-v1", AddedOrderIndexes: []int32{9}, RejectedOrderIndexes: nil, SkippedExisting: 2, MaxExistingOrderIndex: 7},
	}, got)
}

func Test_WorkbookSeeder_SeedWorkbooks_shouldReportRejected_whenQuestionRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: nothing imported yet; the server rejects C1
	client := existingVocabularyClient(ctx, t)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C1")).
		Return(errQuestionRejected)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C2")).Return(nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C3")).Return(nil)
	seeder := seed.NewWorkbookSeeder(client, existingVocabularyEnsurer(ctx, t), discardLogger())

	// when
	got, err := seeder.SeedWorkbooks(ctx, testOrgID, testPublicSpaceID, csvSeeds())

	// then
	require.NoError(t, err)
	assert.Equal(t, []seed.WorkbookOutput{
		{SeedKey: "vocab-v1", AddedOrderIndexes: []int32{7, 9}, RejectedOrderIndexes: []int32{5}, SkippedExisting: 0, MaxExistingOrderIndex: 0},
	}, got)
}

func Test_WorkbookSeeder_SeedWorkbooks_shouldLogToInjectedLogger_whenWorkbookAlreadySeeded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{{WorkbookID: "wb-existing", Description: "[seedKey:vocab-v1]"}}, nil)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	seeder := seed.NewWorkbookSeeder(client, existingVocabularyEnsurer(ctx, t), logger)
	seeds := []seed.PublicWorkbookSeed{{SeedKey: "vocab-v1", Title: "Vocabulary", Language: "ja"}}

	// when
	_, err := seeder.SeedWorkbooks(ctx, testOrgID, testPublicSpaceID, seeds)

	// then
	require.NoError(t, err)
	assert.Contains(t, logs.String(), `"msg":"workbook already seeded"`)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldNotReAddExistingQuestions_byTagSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: workbook exists with seedKey marker, and one of its two questions already exists
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{{
			WorkbookID:  "wb-existing",
			Title:       "Vocabulary",
			Description: "vocab desc [seedKey:vocab-v1]",
		}}, nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-existing").
		Return([]seed.QuestionListItem{{
			QuestionID: "q-existing",
			Tags:       []string{"seed-vocab-v1:q1"},
		}}, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(input seed.AddQuestionInput) bool {
		return input.Content == "C2" && assert.Contains(t, input.Tags, "seed-vocab-v1:q2")
	})).Return(nil)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-existing").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: only the missing question (q2) is added
	require.NoError(t, err)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldReturnError_whenListWorkbooksFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listErr := errors.New("auth service unavailable")
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, listErr)

	ensurer := NewMockWorkbookPolicyEnsurer(t)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.ErrorIs(t, err, listErr)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldReturnError_whenCreateWorkbookFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createErr := errors.New("conflict")
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.Anything).
		Return("", createErr)

	ensurer := NewMockWorkbookPolicyEnsurer(t)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.ErrorIs(t, err, createErr)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldReturnErrInvalidSeed_whenLanguageIsNotISO6391(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language string
	}{
		{name: "uppercase", language: "JA"},
		{name: "three letters", language: "jpn"},
		{name: "region subtag", language: "ja-JP"},
		{name: "empty", language: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given: no workbook exists yet, so the seeder would create one
			client := NewMockWorkbookAPIClient(t)
			client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
				Return(nil, nil)
			ensurer := NewMockWorkbookPolicyEnsurer(t)
			seeds := []seed.PublicWorkbookSeed{{SeedKey: "vocab-v1", Title: "Vocabulary", Language: tt.language}}
			seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), seeds)

			// when
			err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

			// then
			require.ErrorIs(t, err, seed.ErrInvalidSeed)
		})
	}
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldReturnError_whenEnsurePoliciesFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: the workbook is created but granting its policies fails
	ensureErr := errors.New("forbidden")
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.Anything).
		Return("wb-Vocabulary", nil)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Vocabulary").
		Return(ensureErr)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: the error propagates and no questions are added
	require.ErrorIs(t, err, ensureErr)
}

// existingVocabularyClient returns a client mock whose vocab-v1 workbook
// exists with no questions yet.
func existingVocabularyClient(ctx context.Context, t *testing.T) *MockWorkbookAPIClient {
	t.Helper()
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return([]seed.WorkbookListItem{{WorkbookID: "wb-existing", Description: "[seedKey:vocab-v1]"}}, nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-existing").Return(nil, nil)
	return client
}

func existingVocabularyEnsurer(ctx context.Context, t *testing.T) *MockWorkbookPolicyEnsurer {
	t.Helper()
	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-existing").Return(nil)
	return ensurer
}

func addQuestionWithContent(content string) any {
	return mock.MatchedBy(func(input seed.AddQuestionInput) bool { return input.Content == content })
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldAddRemainingQuestions_whenQuestionRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: the server rejects C1
	client := existingVocabularyClient(ctx, t)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C1")).
		Return(errQuestionRejected)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C2")).
		Return(nil)
	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, existingVocabularyEnsurer(ctx, t), discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: C2 is still added before the rejection is reported
	require.ErrorIs(t, err, seed.ErrInvalidSeed)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldSeedLaterWorkbooks_whenEarlierWorkbookQuestionRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: vocab-v1's C1 is rejected; grammar-v1 comes after it and must still be created
	client := existingVocabularyClient(ctx, t)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C1")).
		Return(errQuestionRejected)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C2")).
		Return(nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(input seed.CreateWorkbookInput) bool {
		return input.Title == "Grammar"
	})).Return("wb-Grammar", nil)
	ensurer := existingVocabularyEnsurer(ctx, t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Grammar").Return(nil)
	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: grammar-v1 was created (verified by the mock expectations) and the rejection is still reported
	require.ErrorIs(t, err, seed.ErrInvalidSeed)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldReturnError_whenAddQuestionFailsWithoutRejection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: C1 fails for a reason other than rejection; C2 must not be attempted
	sendErr := errors.New("connection reset")
	client := existingVocabularyClient(ctx, t)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", addQuestionWithContent("C1")).
		Return(sendErr)
	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, existingVocabularyEnsurer(ctx, t), discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.ErrorIs(t, err, sendErr)
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldEmitQuestionTagsMatchingDomainPattern_onAddQuestion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	var captured []seed.AddQuestionInput
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.Anything).
		Return("wb-Vocabulary", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.Anything).
		Run(func(_ context.Context, _, _ string, input seed.AddQuestionInput) {
			captured = append(captured, input)
		}).Return(nil).Times(2)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Vocabulary").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
	require.Len(t, captured, 2)
	for _, input := range captured {
		assert.NotEmpty(t, input.Tags, "seeder must prepend the seed identity tag")
		for _, tag := range input.Tags {
			assert.Regexp(t, questionTagPattern, tag,
				"every emitted tag must satisfy cocotola-question's tag pattern")
		}
	}
}

func Test_PublicWorkbookSeeder_SeedPublicWorkbooks_shouldEmbedSeedKeyMarker_inDescription(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(input seed.CreateWorkbookInput) bool {
		return strings.HasSuffix(input.Description, "[seedKey:vocab-v1]")
	})).Return("wb-Vocabulary", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.Anything).
		Return(nil).Times(2)

	ensurer := NewMockWorkbookPolicyEnsurer(t)
	ensurer.EXPECT().EnsureSystemOwnerWorkbookPolicies(ctx, testOrgID, "wb-Vocabulary").Return(nil)

	seeder := seed.NewPublicWorkbookSeeder(seed.NewWorkbookSeeder(client, ensurer, discardLogger()), sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
}
