package question

import (
	"context"
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// UpdateQuestionCommand handles question updates.
type UpdateQuestionCommand struct {
	questionFinder questionFinder
	questionSaver  questionSaver
	authChecker    authorizationChecker
}

// NewUpdateQuestionCommand returns a new UpdateQuestionCommand.
func NewUpdateQuestionCommand(questionFinder questionFinder, questionSaver questionSaver, authChecker authorizationChecker) *UpdateQuestionCommand {
	return &UpdateQuestionCommand{
		questionFinder: questionFinder,
		questionSaver:  questionSaver,
		authChecker:    authChecker,
	}
}

// UpdateQuestion updates an existing question.
func (c *UpdateQuestionCommand) UpdateQuestion(ctx context.Context, input *questionservice.UpdateQuestionInput) (*questionservice.UpdateQuestionOutput, error) {
	resource, err := domain.ResourceWorkbook(input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("resource workbook: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionUpdateQuestion(), resource)
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	q, err := c.questionFinder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("find question: %w", err)
	}

	now := time.Now()
	if err := q.Edit(input.Content, input.Tags, input.OrderIndex, now); err != nil {
		return nil, fmt.Errorf("edit question: %w", err)
	}

	if err := markAudioPendingIfWordFill(q, now); err != nil {
		return nil, fmt.Errorf("mark audio pending: %w", err)
	}

	if err := c.questionSaver.Save(ctx, q); err != nil {
		return nil, fmt.Errorf("save question: %w", err)
	}

	return &questionservice.UpdateQuestionOutput{Item: toServiceItem(q)}, nil
}
