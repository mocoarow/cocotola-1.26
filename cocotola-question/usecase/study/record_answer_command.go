package study

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainstudy "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

// RecordAnswerCommand handles recording a study answer.
type RecordAnswerCommand struct {
	studyRecordFinder studyRecordFinder
	studyRecordSaver  studyRecordSaver
	activeListRepo    activeQuestionListFinder
	workbookRepo      workbookFinder
	authChecker       authorizationChecker
	config            UsecaseConfig
}

// NewRecordAnswerCommand returns a new RecordAnswerCommand.
func NewRecordAnswerCommand(
	studyRecordFinder studyRecordFinder,
	studyRecordSaver studyRecordSaver,
	activeListRepo activeQuestionListFinder,
	workbookRepo workbookFinder,
	authChecker authorizationChecker,
	config UsecaseConfig,
) *RecordAnswerCommand {
	return &RecordAnswerCommand{
		studyRecordFinder: studyRecordFinder,
		studyRecordSaver:  studyRecordSaver,
		activeListRepo:    activeListRepo,
		workbookRepo:      workbookRepo,
		authChecker:       authChecker,
		config:            config,
	}
}

// RecordAnswer records an answer and updates the study record.
func (c *RecordAnswerCommand) RecordAnswer(ctx context.Context, input *studyservice.RecordAnswerInput) (*studyservice.RecordAnswerOutput, error) {
	wb, err := c.workbookRepo.FindByID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find workbook: %w", err)
	}

	resource, err := domain.ResourceWorkbook(wb.ID())
	if err != nil {
		return nil, fmt.Errorf("resource workbook: %w", err)
	}
	allowed, err := c.authChecker.IsAllowed(ctx, input.OrganizationID, input.OperatorID, domain.ActionStudyWorkbook(), resource)
	if err != nil {
		return nil, fmt.Errorf("authorization check: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	activeList, err := c.activeListRepo.FindByWorkbookID(ctx, input.WorkbookID)
	if err != nil {
		return nil, fmt.Errorf("find active question list: %w", err)
	}
	if !activeList.Contains(input.QuestionID) {
		return nil, fmt.Errorf("question %s not found in workbook %s: %w", input.QuestionID, input.WorkbookID, domain.ErrQuestionNotFound)
	}

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

	now := c.config.Now()
	if input.Correct {
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
