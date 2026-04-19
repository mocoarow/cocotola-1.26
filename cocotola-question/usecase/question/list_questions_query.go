package question

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// ListQuestionsQuery handles listing questions in a workbook.
type ListQuestionsQuery struct {
	questionRepo questionFinder
	workbookRepo workbookFinder
	authChecker  authorizationChecker
}

// NewListQuestionsQuery returns a new ListQuestionsQuery.
func NewListQuestionsQuery(questionRepo questionFinder, workbookRepo workbookFinder, authChecker authorizationChecker) *ListQuestionsQuery {
	return &ListQuestionsQuery{
		questionRepo: questionRepo,
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// ListQuestions returns all questions in a workbook.
func (q *ListQuestionsQuery) ListQuestions(ctx context.Context, input *questionservice.ListQuestionsInput) (*questionservice.ListQuestionsOutput, error) {
	wb, err := q.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if !wb.Visibility().IsPublic() {
		allowed, err := q.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionViewWorkbook(), domain.ResourceWorkbook(input.WorkbookID))
		if err != nil {
			return nil, fmt.Errorf("authorization check: %w", err)
		}
		if !allowed {
			return nil, domain.ErrForbidden
		}
	}

	questions, err := q.questionRepo.FindByWorkbookID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find questions: %w", err)
	}

	items := make([]questionservice.Item, len(questions))
	for i, question := range questions {
		items[i] = questionservice.Item{
			QuestionID:   question.ID(),
			QuestionType: question.QuestionType().Value(),
			Content:      question.Content(),
			Tags:         question.Tags(),
			OrderIndex:   question.OrderIndex(),
			CreatedAt:    question.CreatedAt(),
			UpdatedAt:    question.UpdatedAt(),
		}
	}

	return &questionservice.ListQuestionsOutput{Questions: items}, nil
}
