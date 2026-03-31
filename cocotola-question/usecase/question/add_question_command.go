package question

import (
	"context"
	"fmt"
	"time"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// AddQuestionCommand handles adding a question to a workbook.
type AddQuestionCommand struct {
	questionRepo questionAdder
	workbookRepo workbookFinder
	authChecker  authorizationChecker
}

// NewAddQuestionCommand returns a new AddQuestionCommand.
func NewAddQuestionCommand(questionRepo questionAdder, workbookRepo workbookFinder, authChecker authorizationChecker) *AddQuestionCommand {
	return &AddQuestionCommand{
		questionRepo: questionRepo,
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// AddQuestion adds a question to a workbook.
func (c *AddQuestionCommand) AddQuestion(ctx context.Context, input *questionservice.AddQuestionInput) (*questionservice.AddQuestionOutput, error) {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domainrbac.ActionCreateQuestion(), domainrbac.ResourceAny())
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	wb, err := c.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if wb.OwnerID() != input.OperatorID {
		return nil, domain.ErrForbidden
	}

	if _, err := domainquestion.NewType(input.QuestionType); err != nil {
		return nil, fmt.Errorf("new question type: %w", err)
	}

	questionID, err := c.questionRepo.Add(ctx, input.WorkbookID, input.QuestionType, input.Content, input.OrderIndex)
	if err != nil {
		return nil, fmt.Errorf("add question: %w", err)
	}

	now := time.Now()
	return &questionservice.AddQuestionOutput{
		Item: questionservice.Item{
			QuestionID:   questionID,
			QuestionType: input.QuestionType,
			Content:      input.Content,
			OrderIndex:   input.OrderIndex,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}, nil
}
