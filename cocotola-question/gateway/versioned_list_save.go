package gateway

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
)

// versionedListEntity is the minimal contract every list-style aggregate
// (ActiveQuestionList, OwnedWorkbookList, ...) implements: it holds a
// monotonic version and exposes its current entries so the gateway can
// build the persistence record.
type versionedListEntity[E any] interface {
	Entries() []E
	Version() int
	SetVersion(int)
}

// saveVersionedList centralizes the optimistic-locking save flow shared by
// active_question_list and owned_workbook_list. It converts the aggregate's
// version-bumped entries into a record, hands it to firestoresave, and maps
// the generic "not found" error onto the caller-specified domain error.
//
// The two list repositories used to inline this almost identical block; this
// helper exists so adding a new versioned-list aggregate is a one-liner and
// the lint check stays happy.
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
