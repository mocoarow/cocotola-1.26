package gateway

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
)

// versionedListEntity is the contract shared by list-style aggregates.
type versionedListEntity[E any] interface {
	Entries() []E
	Version() int
	SetVersion(int)
}

// saveVersionedList is the optimistic-locking save shared by list repositories;
// it maps libversioned.ErrNotFound to notFoundErr.
func saveVersionedList[Entity versionedListEntity[E], E any, Record libversioned.Record](
	ctx context.Context,
	client *firestore.Client,
	docRef *firestore.DocumentRef,
	entity Entity,
	newRecord Record,
	decode func(*firestore.DocumentSnapshot) (int, error),
	entityName string,
	notFoundErr error,
) error {
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[Record]{
		Client:     client,
		Entity:     entity,
		DocRef:     docRef,
		NewRecord:  newRecord,
		Decode:     decode,
		EntityName: entityName,
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return notFoundErr
	}
	if err != nil {
		return fmt.Errorf("save %s: %w", entityName, err)
	}
	return nil
}
