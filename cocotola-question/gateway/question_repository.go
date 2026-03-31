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

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

const questionsSubCollection = "questions"

type questionRecord struct {
	QuestionType string    `firestore:"questionType"`
	Content      string    `firestore:"content"`
	OrderIndex   int       `firestore:"orderIndex"`
	CreatedAt    time.Time `firestore:"createdAt"`
	UpdatedAt    time.Time `firestore:"updatedAt"`
}

func toQuestionDomain(id string, r *questionRecord) (*domainquestion.Question, error) {
	qt, err := domainquestion.NewType(r.QuestionType)
	if err != nil {
		return nil, fmt.Errorf("invalid question type %q: %w", r.QuestionType, err)
	}
	return domainquestion.ReconstructQuestion(id, qt, r.Content, r.OrderIndex, r.CreatedAt, r.UpdatedAt), nil
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

// Add inserts a new question and returns the auto-generated document ID.
func (r *QuestionRepository) Add(ctx context.Context, workbookID string, questionType string, content string, orderIndex int) (string, error) {
	now := time.Now()
	record := questionRecord{
		QuestionType: questionType,
		Content:      content,
		OrderIndex:   orderIndex,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	docRef, _, err := r.questionsCol(workbookID).Add(ctx, record)
	if err != nil {
		return "", fmt.Errorf("add question: %w", err)
	}
	return docRef.ID, nil
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
	q, err := toQuestionDomain(doc.Ref.ID, &record)
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
		q, err := toQuestionDomain(doc.Ref.ID, &record)
		if err != nil {
			return nil, fmt.Errorf("convert question domain: %w", err)
		}
		questions = append(questions, *q)
	}
	return questions, nil
}

// Update updates an existing question.
func (r *QuestionRepository) Update(ctx context.Context, workbookID string, questionID string, content string, orderIndex int) error {
	now := time.Now()
	_, err := r.questionsCol(workbookID).Doc(questionID).Set(ctx, map[string]any{
		"content":    content,
		"orderIndex": orderIndex,
		"updatedAt":  now,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	return nil
}

// Delete removes a question document.
func (r *QuestionRepository) Delete(ctx context.Context, workbookID string, questionID string) error {
	_, err := r.questionsCol(workbookID).Doc(questionID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	return nil
}
