package question

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// AddQuestionCommand handles adding a question to a workbook.
type AddQuestionCommand struct {
	questionRepo     questionAdder
	activeListFinder activeQuestionListFinder
	activeListSaver  activeQuestionListSaver
	authChecker      authorizationChecker
}

// NewAddQuestionCommand returns a new AddQuestionCommand.
func NewAddQuestionCommand(questionRepo questionAdder, activeListFinder activeQuestionListFinder, activeListSaver activeQuestionListSaver, authChecker authorizationChecker) *AddQuestionCommand {
	return &AddQuestionCommand{
		questionRepo:     questionRepo,
		activeListFinder: activeListFinder,
		activeListSaver:  activeListSaver,
		authChecker:      authChecker,
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

	if _, err := domainquestion.NewType(input.QuestionType); err != nil {
		return nil, fmt.Errorf("new question type: %w", err)
	}

	questionID, err := c.questionRepo.Add(ctx, input.WorkbookID, input.QuestionType, input.Content, input.Tags, input.OrderIndex)
	if err != nil {
		return nil, fmt.Errorf("add question: %w", err)
	}

	// Add to active question list (eventual consistency).
	if err := c.saveActiveList(ctx, input.WorkbookID, questionID); err != nil {
		slog.ErrorContext(ctx, "active question list save failed after question creation",
			slog.String("question_id", questionID),
			slog.String("workbook_id", input.WorkbookID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("save active question list: %w", err)
	}

	now := time.Now()
	return &questionservice.AddQuestionOutput{
		Item: questionservice.Item{
			QuestionID:   questionID,
			QuestionType: input.QuestionType,
			Content:      input.Content,
			Tags:         input.Tags,
			OrderIndex:   input.OrderIndex,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}, nil
}

func (c *AddQuestionCommand) saveActiveList(ctx context.Context, workbookID string, questionID string) error {
	activeList, err := c.activeListFinder.FindByWorkbookID(ctx, workbookID)
	if err != nil {
		return fmt.Errorf("find active question list: %w", err)
	}
	if err := activeList.Add(questionID); err != nil {
		return fmt.Errorf("add to active question list: %w", err)
	}
	if err := c.activeListSaver.Save(ctx, activeList); err != nil {
		return fmt.Errorf("save active question list: %w", err)
	}
	return nil
}
