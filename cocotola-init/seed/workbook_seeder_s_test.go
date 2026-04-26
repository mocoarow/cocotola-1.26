//go:build small

package seed_test

import (
	"context"
	"errors"
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

func sampleSeeds() []seed.PublicWorkbookSeed {
	return []seed.PublicWorkbookSeed{
		{
			SeedKey:     "vocab-v1",
			Title:       "Vocabulary",
			Description: "vocab desc",
			Questions: []seed.QuestionSeed{
				{SeedKey: "q1", QuestionType: "word_fill", Content: "C1", OrderIndex: 0},
				{SeedKey: "q2", QuestionType: "word_fill", Content: "C2", OrderIndex: 1},
			},
		},
		{
			SeedKey:     "grammar-v1",
			Title:       "Grammar",
			Description: "grammar desc",
		},
	}
}

func Test_WorkbookSeeder_shouldCreateAllWorkbooksAndQuestions_onFirstRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(body seed.CreateWorkbookRequest) bool {
		return body.Title == "Vocabulary" && body.Visibility == "public" &&
			strings.Contains(body.Description, "[seedKey:vocab-v1]")
	})).Return("wb-Vocabulary", nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(body seed.CreateWorkbookRequest) bool {
		return body.Title == "Grammar" && body.Visibility == "public" &&
			strings.Contains(body.Description, "[seedKey:grammar-v1]")
	})).Return("wb-Grammar", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.MatchedBy(func(body seed.AddQuestionRequest) bool {
		return body.Content == "C1"
	})).Return(nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.MatchedBy(func(body seed.AddQuestionRequest) bool {
		return body.Content == "C2"
	})).Return(nil)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
}

func Test_WorkbookSeeder_shouldBeIdempotent_onSecondRun(t *testing.T) {
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

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: no CreateWorkbook or AddQuestion calls
	require.NoError(t, err)
}

func Test_WorkbookSeeder_shouldDetectExistingWorkbook_byDescriptionSeedKey_evenWhenTitleChanges(t *testing.T) {
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
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(body seed.AddQuestionRequest) bool {
		return body.Content == "C1"
	})).Return(nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(body seed.AddQuestionRequest) bool {
		return body.Content == "C2"
	})).Return(nil)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: the workbook is reused (not recreated) and its questions are added
	require.NoError(t, err)
}

func Test_WorkbookSeeder_shouldNotReAddExistingQuestions_byTagSeedKey(t *testing.T) {
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
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-existing", mock.MatchedBy(func(body seed.AddQuestionRequest) bool {
		return body.Content == "C2" && assert.Contains(t, body.Tags, "seed-vocab-v1:q2")
	})).Return(nil)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then: only the missing question (q2) is added
	require.NoError(t, err)
}

func Test_WorkbookSeeder_shouldReturnError_whenListWorkbooksFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listErr := errors.New("auth service unavailable")
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, listErr)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.ErrorIs(t, err, listErr)
}

func Test_WorkbookSeeder_shouldReturnError_whenCreateWorkbookFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createErr := errors.New("conflict")
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.Anything).
		Return("", createErr)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds())

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.ErrorIs(t, err, createErr)
}

func Test_WorkbookSeeder_shouldEmitQuestionTagsMatchingDomainPattern_onAddQuestion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	var captured []seed.AddQuestionRequest
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.Anything).
		Return("wb-Vocabulary", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.Anything).
		Run(func(_ context.Context, _, _ string, body seed.AddQuestionRequest) {
			captured = append(captured, body)
		}).Return(nil).Times(2)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
	require.Len(t, captured, 2)
	for _, body := range captured {
		assert.NotEmpty(t, body.Tags, "seeder must prepend the seed identity tag")
		for _, tag := range body.Tags {
			assert.Regexp(t, questionTagPattern, tag,
				"every emitted tag must satisfy cocotola-question's tag pattern")
		}
	}
}

func Test_WorkbookSeeder_shouldEmbedSeedKeyMarker_inDescription(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	client := NewMockWorkbookAPIClient(t)
	client.EXPECT().ListWorkbooks(ctx, testOrgID, testPublicSpaceID).
		Return(nil, nil)
	client.EXPECT().CreateWorkbook(ctx, testOrgID, mock.MatchedBy(func(body seed.CreateWorkbookRequest) bool {
		return strings.HasSuffix(body.Description, "[seedKey:vocab-v1]")
	})).Return("wb-Vocabulary", nil)
	client.EXPECT().ListQuestions(ctx, testOrgID, "wb-Vocabulary").
		Return(nil, nil)
	client.EXPECT().AddQuestion(ctx, testOrgID, "wb-Vocabulary", mock.Anything).
		Return(nil).Times(2)

	seeder := seed.NewWorkbookSeeder(client, sampleSeeds()[:1])

	// when
	err := seeder.SeedPublicWorkbooks(ctx, testOrgID, testPublicSpaceID)

	// then
	require.NoError(t, err)
}
