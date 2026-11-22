package question

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// DeleteQuestionCommand handles question deletion.
type DeleteQuestionCommand struct {
	questionDeleter  questionDeleter
	workbookFinder   workbookFinder
	activeListFinder activeQuestionListFinder
	activeListSaver  activeQuestionListSaver
	authChecker      authorizationChecker
}

// NewDeleteQuestionCommand returns a new DeleteQuestionCommand.
func NewDeleteQuestionCommand(questionDeleter questionDeleter, workbookFinder workbookFinder, activeListFinder activeQuestionListFinder, activeListSaver activeQuestionListSaver, authChecker authorizationChecker) *DeleteQuestionCommand {
	return &DeleteQuestionCommand{
		questionDeleter:  questionDeleter,
		workbookFinder:   workbookFinder,
		activeListFinder: activeListFinder,
		activeListSaver:  activeListSaver,
		authChecker:      authChecker,
	}
}

// DeleteQuestion deletes a question from a workbook.
func (c *DeleteQuestionCommand) DeleteQuestion(ctx context.Context, input *questionservice.DeleteQuestionInput) error {
	if _, err := c.workbookFinder.FindByID(ctx, input.WorkbookID); err != nil {
		return fmt.Errorf("find workbook: %w", err)
	}

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

	// Remove from active question list (eventual consistency).
	if err := updateActiveList(ctx, c.activeListFinder, c.activeListSaver, input.WorkbookID, func(l *domain.ActiveQuestionList) error {
		l.Remove(input.QuestionID)
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "active question list save failed after question deletion",
			slog.String("question_id", input.QuestionID),
			slog.String("workbook_id", input.WorkbookID),
			slog.Any("error", err),
		)
		return fmt.Errorf("remove question %s from active list: %w", input.QuestionID, err)
	}

	return nil
}
