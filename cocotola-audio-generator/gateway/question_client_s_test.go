package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-audio-generator/gateway"
)

func newTestQuestionClient(t *testing.T, handler http.Handler) *gateway.QuestionAPIClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return gateway.NewQuestionAPIClient(srv.URL, "test-key", 5*time.Second)
}

func Test_QuestionAPIClient_ListPending_shouldReturnItems_whenServerReturns200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/internal/audio/questions/pending", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-Service-Api-Key"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"workbookId":  "wb-1",
					"questionId":  "q-1",
					"sourceText":  "りんご",
					"sourceLang":  "ja",
					"targetText":  "apple",
					"targetLang":  "en",
					"inputHash":   "abc",
					"failedTries": 0,
				},
			},
		})
	})
	client := newTestQuestionClient(t, handler)

	// when
	items, err := client.ListPending(context.Background(), 10)

	// then
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "wb-1", items[0].WorkbookID)
	assert.Equal(t, "q-1", items[0].QuestionID)
	assert.Equal(t, "abc", items[0].InputHash)
}

func Test_QuestionAPIClient_ListPending_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	client := newTestQuestionClient(t, handler)

	// when
	items, err := client.ListPending(context.Background(), 10)

	// then
	require.Error(t, err)
	assert.Nil(t, items)
}

func Test_QuestionAPIClient_Claim_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Claim(context.Background(), item)

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Claim_shouldReturnErrClaimRace_whenServerReturns409(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Claim(context.Background(), item)

	// then
	require.ErrorIs(t, err, domain.ErrClaimRace)
}

func Test_QuestionAPIClient_Claim_shouldReturnError_whenServerReturnsNon200NonConflict(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Claim(context.Background(), item)

	// then
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrClaimRace)
}

func Test_QuestionAPIClient_Complete_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}
	refs := map[string]domain.AudioRef{
		domain.SlotSource: {Path: "path/src.opus", DurationSec: 1.5, SizeBytes: 1024},
	}

	// when
	err := client.Complete(context.Background(), item, refs)

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Complete_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Complete(context.Background(), item, nil)

	// then
	require.Error(t, err)
}

func Test_QuestionAPIClient_Fail_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Fail(context.Background(), item, "tts error")

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Fail_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	client := newTestQuestionClient(t, handler)
	item := domain.PendingItem{WorkbookID: "wb-1", QuestionID: "q-1", InputHash: "hash1"}

	// when
	err := client.Fail(context.Background(), item, "reason")

	// then
	require.Error(t, err)
}

func Test_QuestionAPIClient_ReclaimStale_shouldReturnCount_whenServerReturns200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/internal/audio/questions/reclaim-stale", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"reclaimed": 3})
	})
	client := newTestQuestionClient(t, handler)

	// when
	n, err := client.ReclaimStale(context.Background(), 15*time.Minute, 10)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}

func Test_QuestionAPIClient_ReclaimStale_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()

	// given
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	client := newTestQuestionClient(t, handler)

	// when
	n, err := client.ReclaimStale(context.Background(), 15*time.Minute, 10)

	// then
	require.Error(t, err)
	assert.Equal(t, 0, n)
}
