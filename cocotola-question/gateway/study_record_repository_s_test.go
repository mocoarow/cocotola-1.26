package gateway_test

import (
	"errors"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

// --- mocks ---

type mockStudyRecordIter struct {
	docs []*firestore.DocumentSnapshot
	err  error
	idx  int
}

func (m *mockStudyRecordIter) Next() (*firestore.DocumentSnapshot, error) {
	if m.idx >= len(m.docs) {
		if m.err != nil {
			return nil, m.err
		}
		return nil, iterator.Done
	}
	doc := m.docs[m.idx]
	m.idx++
	return doc, nil
}

func (m *mockStudyRecordIter) Stop() {}

type mockStudyRecordBulkDeleter struct {
	deleteErr error
	jobErrs   []error
	deleted   int
}

func (m *mockStudyRecordBulkDeleter) Delete(_ *firestore.DocumentRef) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}

	m.deleted++

	return nil
}

func (m *mockStudyRecordBulkDeleter) Wait() []error {
	n := min(m.deleted, len(m.jobErrs))
	errs := make([]error, 0, n)
	for _, e := range m.jobErrs[:n] {
		if e != nil {
			errs = append(errs, e)
		}
	}
	return errs
}

// --- tests ---

func Test_deleteStudyRecordDocs_shouldReturnNil_whenNoDocuments(t *testing.T) {
	t.Parallel()

	// given
	iter := &mockStudyRecordIter{}
	bw := &mockStudyRecordBulkDeleter{}

	// when
	err := gateway.DeleteStudyRecordDocs(iter, bw)

	// then
	require.NoError(t, err)
}

func Test_deleteStudyRecordDocs_shouldReturnError_whenIterFails(t *testing.T) {
	t.Parallel()

	// given: iterator fails immediately with a non-Done error
	iter := &mockStudyRecordIter{err: errors.New("firestore unavailable")}
	bw := &mockStudyRecordBulkDeleter{}

	// when
	err := gateway.DeleteStudyRecordDocs(iter, bw)

	// then: error surfaces wrapped under the outer delete-study-records message
	require.ErrorContains(t, err, "delete study records")
	require.ErrorContains(t, err, "iterate study records")
	assert.ErrorContains(t, err, "firestore unavailable")
}

func Test_deleteStudyRecordDocs_shouldReturnError_whenEnqueueFails(t *testing.T) {
	t.Parallel()

	// given: one document exists but BulkWriter rejects the enqueue
	doc := &firestore.DocumentSnapshot{Ref: &firestore.DocumentRef{ID: "doc-1"}}
	iter := &mockStudyRecordIter{docs: []*firestore.DocumentSnapshot{doc}}
	bw := &mockStudyRecordBulkDeleter{deleteErr: errors.New("bw closed")}

	// when
	err := gateway.DeleteStudyRecordDocs(iter, bw)

	// then: loop breaks on first enqueue error; error is surfaced
	require.ErrorContains(t, err, "delete study records")
	require.ErrorContains(t, err, "enqueue delete")
	assert.ErrorContains(t, err, "bw closed")
}

func Test_deleteStudyRecordDocs_shouldReturnError_whenJobFails(t *testing.T) {
	t.Parallel()

	// given: two documents are enqueued successfully but the second job fails
	doc1 := &firestore.DocumentSnapshot{Ref: &firestore.DocumentRef{ID: "doc-1"}}
	doc2 := &firestore.DocumentSnapshot{Ref: &firestore.DocumentRef{ID: "doc-2"}}
	iter := &mockStudyRecordIter{docs: []*firestore.DocumentSnapshot{doc1, doc2}}
	bw := &mockStudyRecordBulkDeleter{
		jobErrs: []error{nil, errors.New("permission denied")},
	}

	// when
	err := gateway.DeleteStudyRecordDocs(iter, bw)

	// then: job error is aggregated and returned
	require.ErrorContains(t, err, "delete study records")
	require.ErrorContains(t, err, "delete study record")
	assert.ErrorContains(t, err, "permission denied")
}
