package question

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type questionFinder interface {
	FindByID(ctx context.Context, workbookID string, questionID string) (*domainquestion.Question, error)
	FindByWorkbookID(ctx context.Context, workbookID string) ([]domainquestion.Question, error)
}

type questionSaver interface {
	Save(ctx context.Context, q *domainquestion.Question) error
}

type questionDeleter interface {
	Delete(ctx context.Context, workbookID string, questionID string) error
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
}

type activeQuestionListFinder interface {
	FindByWorkbookID(ctx context.Context, workbookID string) (*domain.ActiveQuestionList, error)
}

type activeQuestionListSaver interface {
	Save(ctx context.Context, list *domain.ActiveQuestionList) error
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}
