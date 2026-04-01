package question

import (
	"context"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type questionAdder interface {
	Add(ctx context.Context, workbookID string, questionType string, content string, orderIndex int) (string, error)
}

type questionFinder interface {
	FindByID(ctx context.Context, workbookID string, questionID string) (*domainquestion.Question, error)
	FindByWorkbookID(ctx context.Context, workbookID string) ([]domainquestion.Question, error)
}

type questionUpdater interface {
	Update(ctx context.Context, workbookID string, questionID string, content string, orderIndex int) error
}

type questionDeleter interface {
	Delete(ctx context.Context, workbookID string, questionID string) error
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}
