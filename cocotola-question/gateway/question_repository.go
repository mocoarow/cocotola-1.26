package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/firestoresave"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

const questionsSubCollection = "questions"

type questionRecord struct {
	QuestionType string    `firestore:"questionType"`
	Content      string    `firestore:"content"`
	Tags         []string  `firestore:"tags,omitempty"`
	OrderIndex   int       `firestore:"orderIndex"`
	Version      int       `firestore:"version"`
	CreatedAt    time.Time `firestore:"createdAt"`
	UpdatedAt    time.Time `firestore:"updatedAt"`
}

func (r *questionRecord) GetVersion() int {
	return r.Version
}

func toQuestionDomain(id string, workbookID string, r *questionRecord) (*domainquestion.Question, error) {
	qt, err := domainquestion.NewType(r.QuestionType)
	if err != nil {
		return nil, fmt.Errorf("invalid question type %q: %w", r.QuestionType, err)
	}
	return domainquestion.ReconstructQuestion(id, workbookID, qt, r.Content, r.Tags, r.OrderIndex, r.Version, r.CreatedAt, r.UpdatedAt), nil
}

func toQuestionRecord(q *domainquestion.Question, version int) questionRecord {
	tags := q.Tags()
	if tags == nil {
		tags = []string{}
	}
	return questionRecord{
		QuestionType: q.QuestionType().Value(),
		Content:      q.Content(),
		Tags:         tags,
		OrderIndex:   q.OrderIndex(),
		Version:      version,
		CreatedAt:    q.CreatedAt(),
		UpdatedAt:    q.UpdatedAt(),
	}
}

// QuestionRepository manages question persistence as a subcollection of workbooks in Firestore.
type QuestionRepository struct {
	client *firestore.Client
}

// NewQuestionRepository returns a new QuestionRepository.
func NewQuestionRepository(client *firestore.Client) *QuestionRepository {
	return &QuestionRepository{client: client}
}

func (r *QuestionRepository) questionsCol(workbookID string) *firestore.CollectionRef {
	return r.client.Collection(workbooksCollection).Doc(workbookID).Collection(questionsSubCollection)
}

// Save persists a question aggregate. New aggregates (version 0) are inserted at
// the document keyed by q.ID(); loaded aggregates (version > 0) are updated under
// optimistic concurrency control via the version field. The repository updates
// the aggregate's version after a successful persist.
func (r *QuestionRepository) Save(ctx context.Context, q *domainquestion.Question) error {
	docRef := r.questionsCol(q.WorkbookID()).Doc(q.ID())
	record := toQuestionRecord(q, q.Version()+1)
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*questionRecord]{
		Client:    r.client,
		Entity:    q,
		DocRef:    docRef,
		NewRecord: &record,
		Decode: func(snap *firestore.DocumentSnapshot) (int, error) {
			var rec questionRecord
			if err := snap.DataTo(&rec); err != nil {
				return 0, fmt.Errorf("decode question: %w", err)
			}
			return rec.Version, nil
		},
		EntityName: "question",
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrQuestionNotFound
	}
	if err != nil {
		return fmt.Errorf("save question: %w", err)
	}
	return nil
}

// FindByID looks up a question by workbook ID and question ID.
func (r *QuestionRepository) FindByID(ctx context.Context, workbookID string, questionID string) (*domainquestion.Question, error) {
	doc, err := r.questionsCol(workbookID).Doc(questionID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrQuestionNotFound
		}
		return nil, fmt.Errorf("find question by id: %w", err)
	}
	var record questionRecord
	if err := doc.DataTo(&record); err != nil {
		return nil, fmt.Errorf("decode question: %w", err)
	}
	q, err := toQuestionDomain(doc.Ref.ID, workbookID, &record)
	if err != nil {
		return nil, fmt.Errorf("convert question domain: %w", err)
	}
	return q, nil
}

// FindByWorkbookID returns all questions for the given workbook, ordered by orderIndex.
func (r *QuestionRepository) FindByWorkbookID(ctx context.Context, workbookID string) ([]domainquestion.Question, error) {
	iter := r.questionsCol(workbookID).OrderBy("orderIndex", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var questions []domainquestion.Question

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate questions: %w", err)
		}
		var record questionRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, fmt.Errorf("decode question: %w", err)
		}
		q, err := toQuestionDomain(doc.Ref.ID, workbookID, &record)
		if err != nil {
			return nil, fmt.Errorf("convert question domain: %w", err)
		}
		questions = append(questions, *q)
	}
	return questions, nil
}

// FindByIDs returns questions matching the given IDs within a workbook.
// Missing documents are silently skipped.
func (r *QuestionRepository) FindByIDs(ctx context.Context, workbookID string, questionIDs []string) ([]domainquestion.Question, error) {
	if len(questionIDs) == 0 {
		return nil, nil
	}

	refs := make([]*firestore.DocumentRef, len(questionIDs))
	for i, id := range questionIDs {
		refs[i] = r.questionsCol(workbookID).Doc(id)
	}

	docs, err := r.client.GetAll(ctx, refs)
	if err != nil {
		return nil, fmt.Errorf("get questions by ids: %w", err)
	}

	questions := make([]domainquestion.Question, 0, len(docs))
	for _, doc := range docs {
		if !doc.Exists() {
			continue
		}
		var record questionRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, fmt.Errorf("decode question: %w", err)
		}
		q, err := toQuestionDomain(doc.Ref.ID, workbookID, &record)
		if err != nil {
			return nil, fmt.Errorf("convert question domain: %w", err)
		}
		questions = append(questions, *q)
	}
	return questions, nil
}

// Delete removes a question document.
func (r *QuestionRepository) Delete(ctx context.Context, workbookID string, questionID string) error {
	_, err := r.questionsCol(workbookID).Doc(questionID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	return nil
}
