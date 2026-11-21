package question

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// questionCreator persists a new question together with its entry in the
// workbook's active question list.
type questionCreator interface {
	Create(ctx context.Context, q *domainquestion.Question) error
}

// AddQuestionCommand handles adding a question to a workbook.
type AddQuestionCommand struct {
	questionRepo questionCreator
	authChecker  authorizationChecker
	parser       contentParser
}

// NewAddQuestionCommand returns a new AddQuestionCommand.
func NewAddQuestionCommand(questionRepo questionCreator, authChecker authorizationChecker, parser contentParser) *AddQuestionCommand {
	return &AddQuestionCommand{
		questionRepo: questionRepo,
		authChecker:  authChecker,
		parser:       parser,
	}
}

// AddQuestion adds a question to a workbook.
func (c *AddQuestionCommand) AddQuestion(ctx context.Context, input *questionservice.AddQuestionInput) (*questionservice.AddQuestionOutput, error) {
	resource, err := domain.ResourceWorkbook(input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("resource workbook: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionCreateQuestion(), resource)
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	qt, err := domainquestion.NewType(input.QuestionType)
	if err != nil {
		return nil, fmt.Errorf("new question type: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate question id: %w", err)
	}
	now := time.Now()
	q, err := domainquestion.NewQuestion(c.parser, id.String(), input.WorkbookID, qt, input.Content, input.Tags, input.OrderIndex, now, now)
	if err != nil {
		return nil, fmt.Errorf("new question: %w", err)
	}

	if err := markAudioPendingIfWordFill(c.parser, q, now); err != nil {
		return nil, fmt.Errorf("mark audio pending: %w", err)
	}

	if err := c.questionRepo.Create(ctx, q); err != nil {
		return nil, fmt.Errorf("create question in workbook %s: %w", input.WorkbookID, err)
	}

	return &questionservice.AddQuestionOutput{Item: toServiceItem(q)}, nil
}
