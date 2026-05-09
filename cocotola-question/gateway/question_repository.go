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

type audioRefRecord struct {
	Path        string  `firestore:"path"`
	DurationSec float64 `firestore:"durationSec"`
	SizeBytes   int64   `firestore:"sizeBytes"`
}

type audioGenerationRecord struct {
	Status         string                    `firestore:"status"`
	InputHash      string                    `firestore:"inputHash"`
	Refs           map[string]audioRefRecord `firestore:"refs,omitempty"`
	UpdatedAt      time.Time                 `firestore:"updatedAt"`
	FailedAttempts int                       `firestore:"failedAttempts,omitempty"`
	LastError      string                    `firestore:"lastError,omitempty"`
}

type questionRecord struct {
	QuestionType    string                 `firestore:"questionType"`
	Content         string                 `firestore:"content"`
	Tags            []string               `firestore:"tags,omitempty"`
	OrderIndex      int                    `firestore:"orderIndex"`
	Version         int                    `firestore:"version"`
	CreatedAt       time.Time              `firestore:"createdAt"`
	UpdatedAt       time.Time              `firestore:"updatedAt"`
	AudioGeneration *audioGenerationRecord `firestore:"audioGeneration,omitempty"`
}

func (r *questionRecord) GetVersion() int {
	return r.Version
}

func toQuestionDomain(id string, workbookID string, r *questionRecord) (*domainquestion.Question, error) {
	qt, err := domainquestion.NewType(r.QuestionType)
	if err != nil {
		return nil, fmt.Errorf("invalid question type %q: %w", r.QuestionType, err)
	}
	q := domainquestion.ReconstructQuestion(id, workbookID, qt, r.Content, r.Tags, r.OrderIndex, r.Version, r.CreatedAt, r.UpdatedAt)
	if r.AudioGeneration != nil {
		ag, err := audioGenerationRecordToDomain(r.AudioGeneration)
		if err != nil {
			return nil, fmt.Errorf("audio generation: %w", err)
		}
		q.SetAudioGeneration(ag)
	}
	return q, nil
}

func toQuestionRecord(q *domainquestion.Question, version int) questionRecord {
	tags := q.Tags()
	if tags == nil {
		tags = []string{}
	}
	return questionRecord{
		QuestionType:    q.QuestionType().Value(),
		Content:         q.Content(),
		Tags:            tags,
		OrderIndex:      q.OrderIndex(),
		Version:         version,
		CreatedAt:       q.CreatedAt(),
		UpdatedAt:       q.UpdatedAt(),
		AudioGeneration: audioGenerationDomainToRecord(q.AudioGeneration()),
	}
}

func audioGenerationDomainToRecord(ag *domainquestion.AudioGeneration) *audioGenerationRecord {
	if ag == nil {
		return nil
	}
	var refs map[string]audioRefRecord
	domainRefs := ag.Refs()
	if len(domainRefs) > 0 {
		refs = make(map[string]audioRefRecord, len(domainRefs))
		for k, v := range domainRefs {
			refs[k] = audioRefRecord{
				Path:        v.Path(),
				DurationSec: v.DurationSec(),
				SizeBytes:   v.SizeBytes(),
			}
		}
	}
	return &audioGenerationRecord{
		Status:         ag.Status().Value(),
		InputHash:      ag.InputHash(),
		Refs:           refs,
		UpdatedAt:      ag.UpdatedAt(),
		FailedAttempts: ag.FailedAttempts(),
		LastError:      ag.LastError(),
	}
}

func audioGenerationRecordToDomain(r *audioGenerationRecord) (*domainquestion.AudioGeneration, error) {
	status, err := domainquestion.NewAudioGenerationStatus(r.Status)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	var refs map[string]domainquestion.AudioRef
	if len(r.Refs) > 0 {
		refs = make(map[string]domainquestion.AudioRef, len(r.Refs))
		for k, v := range r.Refs {
			ref, err := domainquestion.NewAudioRef(v.Path, v.DurationSec, v.SizeBytes)
			if err != nil {
				return nil, fmt.Errorf("ref %q: %w", k, err)
			}
			refs[k] = ref
		}
	}
	ag, err := domainquestion.NewAudioGeneration(status, r.InputHash, refs, r.UpdatedAt, r.FailedAttempts, r.LastError)
	if err != nil {
		return nil, fmt.Errorf("new audio generation: %w", err)
	}
	return ag, nil
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

// FindPendingAudio returns up to limit questions whose audio generation is in
// the pending state, ordered by audioGeneration.updatedAt ascending so older
// queue entries are processed first. Cross-workbook query via collection group.
func (r *QuestionRepository) FindPendingAudio(ctx context.Context, limit int) ([]domainquestion.Question, error) {
	if limit <= 0 {
		return nil, nil
	}
	iter := r.client.CollectionGroup(questionsSubCollection).
		Where("audioGeneration.status", "==", domainquestion.AudioGenerationStatusPending().Value()).
		OrderBy("audioGeneration.updatedAt", firestore.Asc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var questions []domainquestion.Question

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate pending audio: %w", err)
		}
		workbookID, ok := workbookIDFromQuestionRef(doc.Ref)
		if !ok {
			return nil, fmt.Errorf("question doc %s has unexpected parent path", doc.Ref.Path)
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

// FindStaleGenerating returns up to limit questions whose audio generation has
// been stuck in the generating state with audioGeneration.updatedAt older than
// staleBefore. Used by the reaper to reclaim entries left behind by crashed
// workers (the previous batch claimed an item, then died before completing or
// failing it).
func (r *QuestionRepository) FindStaleGenerating(ctx context.Context, staleBefore time.Time, limit int) ([]domainquestion.Question, error) {
	if limit <= 0 {
		return nil, nil
	}
	iter := r.client.CollectionGroup(questionsSubCollection).
		Where("audioGeneration.status", "==", domainquestion.AudioGenerationStatusGenerating().Value()).
		Where("audioGeneration.updatedAt", "<", staleBefore).
		OrderBy("audioGeneration.updatedAt", firestore.Asc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var questions []domainquestion.Question

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate stale generating: %w", err)
		}
		workbookID, ok := workbookIDFromQuestionRef(doc.Ref)
		if !ok {
			return nil, fmt.Errorf("question doc %s has unexpected parent path", doc.Ref.Path)
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

// workbookIDFromQuestionRef extracts the parent workbook ID from a question
// document path of the form workbooks/{workbookID}/questions/{questionID}.
func workbookIDFromQuestionRef(ref *firestore.DocumentRef) (string, bool) {
	if ref == nil || ref.Parent == nil || ref.Parent.Parent == nil {
		return "", false
	}
	return ref.Parent.Parent.ID, true
}
