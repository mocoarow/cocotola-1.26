package gateway

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const activeQuestionListsCollection = "active_question_lists"

type activeQuestionListRecord struct {
	QuestionIDs []string `firestore:"questionIDs"`
	Version     int      `firestore:"version"`
}

func (r *activeQuestionListRecord) GetVersion() int {
	return r.Version
}

// ActiveQuestionListRepository manages active question list persistence in Firestore.
type ActiveQuestionListRepository struct {
	client *firestore.Client
}

// NewActiveQuestionListRepository returns a new ActiveQuestionListRepository.
func NewActiveQuestionListRepository(client *firestore.Client) *ActiveQuestionListRepository {
	return &ActiveQuestionListRepository{client: client}
}

func (r *ActiveQuestionListRepository) listDoc(workbookID string) *firestore.DocumentRef {
	return r.client.Collection(activeQuestionListsCollection).Doc(workbookID)
}

// FindByWorkbookID returns the active question list for the given workbook.
func (r *ActiveQuestionListRepository) FindByWorkbookID(ctx context.Context, workbookID string) (*domain.ActiveQuestionList, error) {
	snap, err := r.listDoc(workbookID).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("get active question list doc: %w", err)
		}
		// Document does not exist yet; return empty list.
		list, err := domain.NewActiveQuestionList(workbookID, nil)
		if err != nil {
			return nil, fmt.Errorf("new active question list: %w", err)
		}
		return list, nil
	}

	var record activeQuestionListRecord
	if err := snap.DataTo(&record); err != nil {
		return nil, fmt.Errorf("decode active question list: %w", err)
	}

	list, err := domain.NewActiveQuestionList(workbookID, record.QuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active question list: %w", err)
	}
	list.SetVersion(record.Version)
	return list, nil
}

// Save persists the active question list atomically using a Firestore transaction.
// It uses optimistic locking via a version field.
func (r *ActiveQuestionListRepository) Save(ctx context.Context, list *domain.ActiveQuestionList) error {
	record := activeQuestionListRecord{
		QuestionIDs: list.Entries(),
		Version:     list.Version() + 1,
	}
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*activeQuestionListRecord]{
		Client:    r.client,
		Entity:    list,
		DocRef:    r.listDoc(list.WorkbookID()),
		NewRecord: &record,
		Decode: func(snap *firestore.DocumentSnapshot) (int, error) {
			var rec activeQuestionListRecord
			if err := snap.DataTo(&rec); err != nil {
				return 0, fmt.Errorf("decode active question list: %w", err)
			}
			return rec.Version, nil
		},
		EntityName: "active question list",
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrActiveQuestionListNotFound
	}
	if err != nil {
		return fmt.Errorf("save active question list: %w", err)
	}
	return nil
}
