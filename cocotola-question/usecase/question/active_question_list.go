package question

import (
	"context"
	"errors"
	"fmt"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const maxActiveListSaveAttempts = 5

func updateActiveList(
	ctx context.Context,
	finder activeQuestionListFinder,
	saver activeQuestionListSaver,
	workbookID string,
	mutate func(*domain.ActiveQuestionList) error,
) error {
	var err error
	for range maxActiveListSaveAttempts {
		err = findMutateSave(ctx, finder, saver, workbookID, mutate)
		if !errors.Is(err, libversioned.ErrConcurrentModification) {
			return err
		}
	}
	return err
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
