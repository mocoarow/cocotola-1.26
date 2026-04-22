package question

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// DeleteQuestionCommand handles question deletion.
type DeleteQuestionCommand struct {
	questionDeleter questionDeleter
	authChecker     authorizationChecker
}

// NewDeleteQuestionCommand returns a new DeleteQuestionCommand.
func NewDeleteQuestionCommand(questionDeleter questionDeleter, authChecker authorizationChecker) *DeleteQuestionCommand {
	return &DeleteQuestionCommand{
		questionDeleter: questionDeleter,
		authChecker:     authChecker,
	}
}

// DeleteQuestion deletes a question from a workbook.
func (c *DeleteQuestionCommand) DeleteQuestion(ctx context.Context, input *questionservice.DeleteQuestionInput) error {
	resource, err := domain.ResourceWorkbook(input.WorkbookID)
	if err != nil {
		return fmt.Errorf("resource workbook: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionDeleteQuestion(), resource)
	if err != nil {
		return fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return domain.ErrForbidden
	}

	if err := c.questionDeleter.Delete(ctx, input.WorkbookID, input.QuestionID); err != nil {
		return fmt.Errorf("delete question: %w", err)
	}

	return nil
}
