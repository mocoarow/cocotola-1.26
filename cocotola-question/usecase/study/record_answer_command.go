package study

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// RecordAnswerCommand handles recording a study answer.
type RecordAnswerCommand struct {
	studyRecordFinder studyRecordFinder
	studyRecordSaver  studyRecordSaver
	activeListRepo    activeQuestionListFinder
	questionFinder    questionFinder
	workbookRepo      workbookFinder
	authChecker       authorizationChecker
	config            UsecaseConfig
}

// NewRecordAnswerCommand returns a new RecordAnswerCommand.
func NewRecordAnswerCommand(
	studyRecordFinder studyRecordFinder,
	studyRecordSaver studyRecordSaver,
	activeListRepo activeQuestionListFinder,
	questionFinder questionFinder,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
	config UsecaseConfig,
) *RecordAnswerCommand {
	return &RecordAnswerCommand{
		studyRecordFinder: studyRecordFinder,
		studyRecordSaver:  studyRecordSaver,
		activeListRepo:    activeListRepo,
		questionFinder:    questionFinder,
		workbookRepo:      workbookRepo,
		authChecker:       authChecker,
		config:            config,
	}
}

func (c *RecordAnswerCommand) checkStudyAuthorization(ctx context.Context, input *studyservice.RecordAnswerInput, wb *domainworkbook.Workbook) error {
	if wb.Visibility().IsPublic() {
		return nil
	}

	resource, err := domain.ResourceWorkbook(wb.ID())
	if err != nil {
		return fmt.Errorf("resource workbook: %w", err)
	}

	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionStudyWorkbook(), resource)
	if err != nil {
		return fmt.Errorf("authorization check: %w", err)
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return nil
}

// evaluateAnswer dispatches per question type to compute correctness from the input.
// For multiple_choice the server is authoritative (set equality against IsCorrect choices);
// for word_fill the client-supplied boolean is trusted as today.
// Mismatched type/payload combinations return ErrInvalidArgument.
func evaluateAnswer(q *domainquestion.Question, input *studyservice.RecordAnswerInput) (bool, error) {
	switch q.QuestionType().Value() {
	case domainquestion.TypeMultipleChoice().Value():
		if input.SelectedChoiceIDs == nil || input.Correct != nil {
			return false, fmt.Errorf("multiple_choice requires selectedChoiceIds only: %w", domain.ErrInvalidArgument)
		}
		mc, err := domainquestion.ParseMultipleChoiceContent(q.Content())
		if err != nil {
			return false, fmt.Errorf("parse multiple_choice content: %w", err)
		}
		ok, err := mc.EvaluateAnswer(*input.SelectedChoiceIDs)
		if err != nil {
			return false, fmt.Errorf("evaluate multiple_choice answer: %w", err)
		}
		return ok, nil
	case domainquestion.TypeWordFill().Value():
		if input.Correct == nil || input.SelectedChoiceIDs != nil {
			return false, fmt.Errorf("word_fill requires correct only: %w", domain.ErrInvalidArgument)
		}
		return *input.Correct, nil
	default:
		return false, fmt.Errorf("unsupported question type %q: %w", q.QuestionType().Value(), domain.ErrInvalidArgument)
	}
}

func (c *RecordAnswerCommand) findOrCreateRecord(ctx context.Context, input *studyservice.RecordAnswerInput) (*domainstudy.Record, error) {
	record, err := c.studyRecordFinder.FindByID(ctx, input.OperatorID, input.WorkbookID, input.QuestionID)
	if err != nil {
		if !errors.Is(err, domain.ErrStudyRecordNotFound) {
			return nil, fmt.Errorf("find study record: %w", err)
		}

		newRecord, err := domainstudy.NewRecord(input.WorkbookID, input.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("new study record: %w", err)
		}

		record = newRecord
	}

	return record, nil
}

// RecordAnswer records an answer and updates the study record.
func (c *RecordAnswerCommand) RecordAnswer(ctx context.Context, input *studyservice.RecordAnswerInput) (*studyservice.RecordAnswerOutput, error) {
	wb, err := c.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	if err := c.checkStudyAuthorization(ctx, input, wb); err != nil {
		return nil, fmt.Errorf("check study authorization: %w", err)
	}

	activeList, err := c.activeListRepo.FindByWorkbookID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find active question list: %w", err)
	}
	if !activeList.Contains(input.QuestionID) {
		return nil, fmt.Errorf("question %s not found in workbook %s: %w", input.QuestionID, input.WorkbookID, domain.ErrQuestionNotFound)
	}

	q, err := c.questionFinder.FindByID(ctx, input.WorkbookID, input.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("find question: %w", err)
	}

	correct, err := evaluateAnswer(q, input)
	if err != nil {
		return nil, fmt.Errorf("evaluate answer: %w", err)
	}

	record, err := c.findOrCreateRecord(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("find or create record: %w", err)
	}

	now := c.config.Now()
	if correct {
		record.RecordCorrect(now)
	} else {
		record.RecordIncorrect(now)
	}

	if err := c.studyRecordSaver.Save(ctx, input.OperatorID, record); err != nil {
		return nil, fmt.Errorf("save study record: %w", err)
	}

	return &studyservice.RecordAnswerOutput{
		NextDueAt:          record.NextDueAt(),
		ConsecutiveCorrect: record.ConsecutiveCorrect(),
		TotalCorrect:       record.TotalCorrect(),
		TotalIncorrect:     record.TotalIncorrect(),
	}, nil
}
