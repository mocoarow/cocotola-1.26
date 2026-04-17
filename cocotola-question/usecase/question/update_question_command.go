package question

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// UpdateQuestionCommand handles question updates.
type UpdateQuestionCommand struct {
	questionFinder  questionFinder
	questionUpdater questionUpdater
	workbookRepo    workbookFinder
	authChecker     authorizationChecker
}

// NewUpdateQuestionCommand returns a new UpdateQuestionCommand.
func NewUpdateQuestionCommand(questionFinder questionFinder, questionUpdater questionUpdater, workbookRepo workbookFinder, authChecker authorizationChecker) *UpdateQuestionCommand {
	return &UpdateQuestionCommand{
		questionFinder:  questionFinder,
		questionUpdater: questionUpdater,
		workbookRepo:    workbookRepo,
		authChecker:     authChecker,
	}
}

// UpdateQuestion updates an existing question.
func (c *UpdateQuestionCommand) UpdateQuestion(ctx context.Context, input *questionservice.UpdateQuestionInput) (*questionservice.UpdateQuestionOutput, error) {
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionUpdateQuestion(), domain.ResourceAny())
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

	q, err := c.questionFinder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("find question: %w", err)
	}

	if err := c.questionUpdater.Update(ctx, input.WorkbookID, input.QuestionID, input.Content, input.Tags, input.OrderIndex); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}

	return &questionservice.UpdateQuestionOutput{
		Item: questionservice.Item{
			QuestionID:   q.ID(),
			QuestionType: q.QuestionType().Value(),
			Content:      input.Content,
			Tags:         input.Tags,
			OrderIndex:   input.OrderIndex,
			CreatedAt:    q.CreatedAt(),
			UpdatedAt:    q.UpdatedAt(),
		},
	}, nil
}
