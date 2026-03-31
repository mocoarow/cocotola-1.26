package sharing

import (
	"context"
	"fmt"

	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

// UnshareCommand handles removing a workbook reference.
type UnshareCommand struct {
	referenceRepo referenceDeleter
}

// NewUnshareCommand returns a new UnshareCommand.
func NewUnshareCommand(referenceRepo referenceDeleter) *UnshareCommand {
	return &UnshareCommand{
		referenceRepo: referenceRepo,
	}
}

// Unshare removes a workbook reference.
func (c *UnshareCommand) Unshare(ctx context.Context, input *referenceservice.UnshareInput) error {
	if err := c.referenceRepo.Delete(ctx, input.OperatorID, input.ReferenceID); err != nil {
		return fmt.Errorf("delete reference: %w", err)
	}
	return nil
}
