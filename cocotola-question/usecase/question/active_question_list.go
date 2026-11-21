package question

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func updateActiveList(
	ctx context.Context,
	finder activeQuestionListFinder,
	saver activeQuestionListSaver,
	workbookID string,
	mutate func(*domain.ActiveQuestionList) error,
) error {
	if err := occ.Retry(func() error {
		return findMutateSave(ctx, finder, saver, workbookID, mutate)
	}); err != nil {
		return fmt.Errorf("update active list of workbook %s: %w", workbookID, err)
	}
	return nil
}

func findMutateSave(
	ctx context.Context,
	finder activeQuestionListFinder,
	saver activeQuestionListSaver,
	workbookID string,
	mutate func(*domain.ActiveQuestionList) error,
) error {
	activeList, err := finder.FindByWorkbookID(ctx, workbookID)
	if err != nil {
		return fmt.Errorf("find active question list: %w", err)
	}
	if err := mutate(activeList); err != nil {
		return fmt.Errorf("update active question list: %w", err)
	}
	if err := saver.Save(ctx, activeList); err != nil {
		return fmt.Errorf("save active question list: %w", err)
	}
	return nil
}
