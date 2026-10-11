package firestoresave

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

// DecodeVersion extracts the persisted version from a Firestore document snapshot.
type DecodeVersion func(*firestore.DocumentSnapshot) (int, error)

// SaveArgs bundles the arguments to SaveVersioned.
type SaveArgs[R versioned.Record] struct {
	Client *firestore.Client
	Entity versioned.Entity
	DocRef *firestore.DocumentRef
	// NewRecord must already carry version Entity.Version()+1.
	NewRecord R
	Decode    DecodeVersion
	// EntityName is used only in error messages.
	EntityName string
}

// SaveVersioned writes NewRecord in a transaction only if the stored version
// equals Entity.Version(). A fresh entity (version 0) never overwrites an
// existing document, even one whose stored version is 0.
func SaveVersioned[R versioned.Record](ctx context.Context, args SaveArgs[R]) error {
	nextVersion := args.Entity.Version() + 1
	if got := args.NewRecord.GetVersion(); got != nextVersion {
		return fmt.Errorf("save %s: record Version=%d does not match expected next version %d", args.EntityName, got, nextVersion)
	}

	if err := args.Client.RunTransaction(ctx, func(_ context.Context, tx *firestore.Transaction) error {
		return saveInTx(tx, args)
	}); err != nil {
		return fmt.Errorf("save %s: %w", args.EntityName, err)
	}

	args.Entity.SetVersion(nextVersion)
	return nil
}

func saveInTx[R versioned.Record](tx *firestore.Transaction, args SaveArgs[R]) error {
	snap, err := tx.Get(args.DocRef)
	if status.Code(err) == codes.NotFound {
		if args.Entity.Version() != 0 {
			return versioned.ErrNotFound
		}
		return setDoc(tx, args)
	}
	if err != nil {
		return fmt.Errorf("get %s doc in tx: %w", args.EntityName, err)
	}

	if args.Entity.Version() == 0 {
		return versioned.ErrConcurrentModification
	}

	currentVersion, err := args.Decode(snap)
	if err != nil {
		return fmt.Errorf("decode %s: %w", args.EntityName, err)
	}
	if currentVersion != args.Entity.Version() {
		return versioned.ErrConcurrentModification
	}
	return setDoc(tx, args)
}

func setDoc[R versioned.Record](tx *firestore.Transaction, args SaveArgs[R]) error {
	if err := tx.Set(args.DocRef, args.NewRecord); err != nil {
		return fmt.Errorf("set %s doc: %w", args.EntityName, err)
	}
	return nil
}
