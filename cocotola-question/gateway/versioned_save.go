package gateway

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

type versionedEntity interface {
	Version() int
	SetVersion(int)
}

// saveVersionedEntity runs a Firestore transaction that performs an
// optimistic-lock check and persists newRecord at docRef.
func saveVersionedEntity(
	ctx context.Context,
	client *firestore.Client,
	entity versionedEntity,
	docRef *firestore.DocumentRef,
	newRecord any,
	decode func(*firestore.DocumentSnapshot) (int, error),
	entityName string,
) error {
	nextVersion := entity.Version() + 1

	if err := client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(docRef)
		currentVersion := 0
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("get %s doc in tx: %w", entityName, err)
			}
		} else {
			currentVersion, err = decode(snap)
			if err != nil {
				return err
			}
		}

		if currentVersion != entity.Version() {
			return domain.ErrConcurrentModification
		}

		if err := tx.Set(docRef, newRecord); err != nil {
			return fmt.Errorf("save %s: %w", entityName, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("run transaction: %w", err)
	}

	entity.SetVersion(nextVersion)

	return nil
}
