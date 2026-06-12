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
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/study"
)

const studyRecordsSubCollection = "study_records"

// studyRecordIter is the subset of *firestore.DocumentIterator used for deletion.
type studyRecordIter interface {
	Next() (*firestore.DocumentSnapshot, error)
	Stop()
}

// deleteJobResult is the subset of *firestore.BulkWriterJob used for result checking.
type deleteJobResult interface {
	Results() ([]*firestore.WriteResult, error)
}

// studyRecordBulkDeleter is the subset of *firestore.BulkWriter used for deletion.
type studyRecordBulkDeleter interface {
	Delete(dr *firestore.DocumentRef) (deleteJobResult, error)
	End()
}

// bulkWriterAdapter adapts *firestore.BulkWriter to studyRecordBulkDeleter.
type bulkWriterAdapter struct {
	bw *firestore.BulkWriter
}

func (a *bulkWriterAdapter) Delete(dr *firestore.DocumentRef) (deleteJobResult, error) {
	return a.bw.Delete(dr)
}

func (a *bulkWriterAdapter) End() {
	a.bw.End()
}

type studyRecordRecord struct {
	WorkbookID         string    `firestore:"workbookID"`
	QuestionID         string    `firestore:"questionID"`
	ConsecutiveCorrect int       `firestore:"consecutiveCorrect"`
	LastAnsweredAt     time.Time `firestore:"lastAnsweredAt"`
	NextDueAt          time.Time `firestore:"nextDueAt"`
	TotalCorrect       int       `firestore:"totalCorrect"`
	TotalIncorrect     int       `firestore:"totalIncorrect"`
	Version            int       `firestore:"version"`
}

func (r *studyRecordRecord) GetVersion() int {
	return r.Version
}

func studyRecordDocID(workbookID string, questionID string) string {
	return workbookID + "__" + questionID
}

// StudyRecordRepository manages study record persistence in Firestore.
type StudyRecordRepository struct {
	client *firestore.Client
}

// NewStudyRecordRepository returns a new StudyRecordRepository.
func NewStudyRecordRepository(client *firestore.Client) *StudyRecordRepository {
	return &StudyRecordRepository{client: client}
}

func (r *StudyRecordRepository) recordsCol(userID string) *firestore.CollectionRef {
	return r.client.Collection(usersCollection).Doc(userID).Collection(studyRecordsSubCollection)
}

// Save persists a study record atomically using a Firestore transaction.
// It uses optimistic locking via a version field.
func (r *StudyRecordRepository) Save(ctx context.Context, userID string, record *study.Record) error {
	docID := studyRecordDocID(record.WorkbookID(), record.QuestionID())
	rec := studyRecordRecord{
		WorkbookID:         record.WorkbookID(),
		QuestionID:         record.QuestionID(),
		ConsecutiveCorrect: record.ConsecutiveCorrect(),
		LastAnsweredAt:     record.LastAnsweredAt(),
		NextDueAt:          record.NextDueAt(),
		TotalCorrect:       record.TotalCorrect(),
		TotalIncorrect:     record.TotalIncorrect(),
		Version:            record.Version() + 1,
	}
	err := firestoresave.SaveVersioned(ctx, firestoresave.SaveArgs[*studyRecordRecord]{
		Client:    r.client,
		Entity:    record,
		DocRef:    r.recordsCol(userID).Doc(docID),
		NewRecord: &rec,
		Decode: func(snap *firestore.DocumentSnapshot) (int, error) {
			var r studyRecordRecord
			if err := snap.DataTo(&r); err != nil {
				return 0, fmt.Errorf("decode study record: %w", err)
			}
			return r.Version, nil
		},
		EntityName: "study record",
	})
	if errors.Is(err, libversioned.ErrNotFound) {
		return domain.ErrStudyRecordNotFound
	}
	if err != nil {
		return fmt.Errorf("save study record: %w", err)
	}
	return nil
}

// FindByID looks up a study record by user, workbook, and question IDs.
func (r *StudyRecordRepository) FindByID(ctx context.Context, userID string, workbookID string, questionID string) (*study.Record, error) {
	docID := studyRecordDocID(workbookID, questionID)
	doc, err := r.recordsCol(userID).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrStudyRecordNotFound
		}
		return nil, fmt.Errorf("find study record: %w", err)
	}
	var rec studyRecordRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("decode study record: %w", err)
	}
	result := study.ReconstructRecord(
		rec.WorkbookID,
		rec.QuestionID,
		rec.ConsecutiveCorrect,
		rec.LastAnsweredAt,
		rec.NextDueAt,
		rec.TotalCorrect,
		rec.TotalIncorrect,
	)
	result.SetVersion(rec.Version)
	return result, nil
}

// DeleteByWorkbookID removes every study record the user has for the given
// workbook. Idempotent: succeeds with zero deletions when there are no records.
// Deletions are issued in parallel via Firestore's BulkWriter. Because the
// operation is not transactional, any iterator/enqueue failure as well as
// per-delete failures are drained and aggregated via errors.Join so partial
// failures are fully reported to the caller.
func (r *StudyRecordRepository) DeleteByWorkbookID(ctx context.Context, userID string, workbookID string) error {
	iter := r.recordsCol(userID).Where("workbookID", "==", workbookID).Documents(ctx)
	bw := &bulkWriterAdapter{bw: r.client.BulkWriter(ctx)}
	return deleteStudyRecordDocs(iter, bw)
}

// deleteStudyRecordDocs drains iter, enqueues deletes via bw, and aggregates
// all iterator, enqueue, and job-level errors via errors.Join.
func deleteStudyRecordDocs(iter studyRecordIter, bw studyRecordBulkDeleter) error {
	defer iter.Stop()

	var (
		jobs    []deleteJobResult
		iterErr error
	)

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			iterErr = fmt.Errorf("iterate study records: %w", err)
			break
		}
		job, err := bw.Delete(doc.Ref)
		if err != nil {
			iterErr = fmt.Errorf("enqueue delete %s: %w", doc.Ref.ID, err)
			break
		}
		jobs = append(jobs, job)
	}
	bw.End()

	jobErrs := make([]error, 0, len(jobs))
	for _, job := range jobs {
		if _, err := job.Results(); err != nil {
			jobErrs = append(jobErrs, fmt.Errorf("delete study record: %w", err))
		}
	}

	if iterErr != nil || len(jobErrs) > 0 {
		return fmt.Errorf("delete study records: %w", errors.Join(append([]error{iterErr}, jobErrs...)...))
	}
	return nil
}

// FindByWorkbookID returns all study records for a user and workbook.
func (r *StudyRecordRepository) FindByWorkbookID(ctx context.Context, userID string, workbookID string) ([]study.Record, error) {
	iter := r.recordsCol(userID).Where("workbookID", "==", workbookID).Documents(ctx)
	defer iter.Stop()

	var records []study.Record

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("iterate study records: %w", err)
		}
		var rec studyRecordRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("decode study record: %w", err)
		}
		result := study.ReconstructRecord(
			rec.WorkbookID,
			rec.QuestionID,
			rec.ConsecutiveCorrect,
			rec.LastAnsweredAt,
			rec.NextDueAt,
			rec.TotalCorrect,
			rec.TotalIncorrect,
		)
		result.SetVersion(rec.Version)
		records = append(records, *result)
	}
	return records, nil
}
