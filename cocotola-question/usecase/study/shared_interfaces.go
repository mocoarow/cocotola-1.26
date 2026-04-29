package study

import (
	"context"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
)

type studyRecordFinder interface {
	FindByID(ctx context.Context, userID string, workbookID string, questionID string) (*domainstudy.Record, error)
	FindByWorkbookID(ctx context.Context, userID string, workbookID string) ([]domainstudy.Record, error)
}

type studyRecordSaver interface {
	Save(ctx context.Context, userID string, record *domainstudy.Record) error
}

type activeQuestionListFinder interface {
	FindByWorkbookID(ctx context.Context, workbookID string) (*domain.ActiveQuestionList, error)
}

type questionBatchFinder interface {
	FindByIDs(ctx context.Context, workbookID string, questionIDs []string) ([]domainquestion.Question, error)
}

type questionFinder interface {
	FindByID(ctx context.Context, workbookID string, questionID string) (*domainquestion.Question, error)
}

type workbookFinder interface {
	FindByID(ctx context.Context, id string) (*domainworkbook.Workbook, error)
}

type authorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}
