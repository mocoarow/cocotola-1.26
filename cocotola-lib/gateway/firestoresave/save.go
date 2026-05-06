package firestoresave

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

// DecodeVersion extracts the persisted version from a Firestore document
// snapshot. Implementations typically decode the snapshot into a record
// struct and return the value of its version field.
type DecodeVersion func(*firestore.DocumentSnapshot) (int, error)

// SaveArgs bundles the arguments to SaveVersioned. R is the persistence record
// type (typically a pointer to a Firestore-tagged struct) constrained by
// versioned.Record so the helper can verify the record's version without
// reflection.
type SaveArgs[R versioned.Record] struct {
	// Client is the Firestore client used for the transaction.
	Client *firestore.Client
	// Entity is the aggregate whose version drives the CAS.
	Entity versioned.Entity
	// DocRef is the document to write.
	DocRef *firestore.DocumentRef
	// NewRecord is the document body to set. The caller must set
	// NewRecord.GetVersion() to Entity.Version()+1 before calling
	// SaveVersioned.
	NewRecord R
	// Decode reads the version from an existing document snapshot for the
	// CAS comparison.
	Decode DecodeVersion
	// EntityName is interpolated into error messages (e.g. "question").
	EntityName string
}

// SaveVersioned runs a Firestore transaction that performs an optimistic-lock
// check and persists NewRecord at DocRef.
//
// The transaction reads the current document, decodes its version via Decode,
// and writes NewRecord only if the persisted version matches Entity.Version().
// On a CAS miss the helper distinguishes:
//
//   - If Entity.Version() == 0 and the document does not exist, it inserts
//     NewRecord (the expected fresh-aggregate path).
//   - If Entity.Version() == 0 and the document exists, it returns
//     versioned.ErrConcurrentModification (someone else inserted first).
//     The helper refuses to overwrite an existing document on the insert path
//     even if its stored version is 0; a fresh aggregate must never replace
//     a record it did not load.
//   - If Entity.Version() > 0 and the document does not exist, it returns
//     versioned.ErrNotFound (the document was deleted; reload will not help).
//     Callers should translate this into their domain-specific not-found error.
//   - If the document exists but its version differs from Entity.Version(),
//     it returns versioned.ErrConcurrentModification (caller should reload
//     and retry).
//
// The caller MUST set the record's version to Entity.Version()+1 before
// calling this helper. The helper verifies this via the Record interface and
// returns an error if it does not match; this guards against a forgotten
// version assignment silently writing a stale version. The helper does not
// mutate fields on NewRecord. On success the helper calls Entity.SetVersion
// with Entity.Version()+1.
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

	// Doc exists. A fresh aggregate (Version()==0) must never overwrite an
	// existing document, regardless of its stored version, so we surface the
	// collision before consulting the version.
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
