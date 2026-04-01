package question

import (
	"context"
	"fmt"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

// GetQuestionQuery handles question retrieval.
type GetQuestionQuery struct {
	questionRepo questionFinder
	workbookRepo workbookFinder
	authChecker  authorizationChecker
}

// NewGetQuestionQuery returns a new GetQuestionQuery.
func NewGetQuestionQuery(questionRepo questionFinder, workbookRepo workbookFinder, authChecker authorizationChecker) *GetQuestionQuery {
	return &GetQuestionQuery{
		questionRepo: questionRepo,
		workbookRepo: workbookRepo,
		authChecker:  authChecker,
	}
}

// GetQuestion retrieves a question by ID.
func (q *GetQuestionQuery) GetQuestion(ctx context.Context, input *questionservice.GetQuestionInput) (*questionservice.GetQuestionOutput, error) {
	wb, err := q.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if !wb.Visibility().IsPublic() {
		allowed, err := q.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domainrbac.ActionViewWorkbook(), domainrbac.ResourceAny())
		if err != nil {
			return nil, fmt.Errorf("authorization check: %w", err)
		}
		if !allowed {
			return nil, domain.ErrForbidden
		}
	}

	question, err := q.questionRepo.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("find question: %w", err)
	}

	return &questionservice.GetQuestionOutput{
		Item: questionservice.Item{
			QuestionID:   question.ID(),
			QuestionType: question.QuestionType().Value(),
			Content:      question.Content(),
			OrderIndex:   question.OrderIndex(),
			CreatedAt:    question.CreatedAt(),
			UpdatedAt:    question.UpdatedAt(),
		},
	}, nil
}
