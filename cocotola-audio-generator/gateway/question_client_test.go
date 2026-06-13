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

func newTestClient(baseURL string) *gateway.QuestionAPIClient {
	return gateway.NewQuestionAPIClient(baseURL, "test-key", 5*time.Second)
}

func samplePendingItem() domain.PendingItem {
	return domain.PendingItem{
		WorkbookID: "wb-1",
		QuestionID: "q-1",
		SourceText: "りんご",
		SourceLang: "ja",
		TargetText: "apple",
		TargetLang: "en",
		InputHash:  "hash-abc",
	}
}

// ListPending

func Test_QuestionAPIClient_ListPending_shouldReturnItems_whenServerReturns200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
					"inputHash":   "hash-abc",
					"failedTries": 0,
				},
			},
		})
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	items, err := client.ListPending(ctx, 10)

	// then
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "wb-1", items[0].WorkbookID)
	assert.Equal(t, "q-1", items[0].QuestionID)
	assert.Equal(t, "hash-abc", items[0].InputHash)
}

func Test_QuestionAPIClient_ListPending_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	items, err := client.ListPending(ctx, 10)

	// then
	require.ErrorContains(t, err, "list pending")
	assert.Nil(t, items)
}

// Claim

func Test_QuestionAPIClient_Claim_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Claim(ctx, samplePendingItem())

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Claim_shouldReturnErrClaimRace_whenServerReturns409(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Claim(ctx, samplePendingItem())

	// then
	require.ErrorIs(t, err, domain.ErrClaimRace)
}

func Test_QuestionAPIClient_Claim_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Claim(ctx, samplePendingItem())

	// then
	require.ErrorContains(t, err, "claim audio")
	assert.NotErrorIs(t, err, domain.ErrClaimRace)
}

// Complete

func Test_QuestionAPIClient_Complete_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)
	refs := map[string]domain.AudioRef{
		domain.SlotSource: {Path: "gs://bucket/src.opus", DurationSec: 1.2, SizeBytes: 4096},
		domain.SlotTarget: {Path: "gs://bucket/tgt.opus", DurationSec: 0.8, SizeBytes: 2048},
	}

	// when
	err := client.Complete(ctx, samplePendingItem(), refs)

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Complete_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Complete(ctx, samplePendingItem(), nil)

	// then
	require.ErrorContains(t, err, "complete audio")
}

// Fail

func Test_QuestionAPIClient_Fail_shouldReturnNil_whenServerReturns200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Fail(ctx, samplePendingItem(), "synth failed")

	// then
	require.NoError(t, err)
}

func Test_QuestionAPIClient_Fail_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	err := client.Fail(ctx, samplePendingItem(), "synth failed")

	// then
	require.ErrorContains(t, err, "fail audio")
}

// ReclaimStale

func Test_QuestionAPIClient_ReclaimStale_shouldReturnReclaimedCount_whenServerReturns200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"reclaimed": 3})
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	count, err := client.ReclaimStale(ctx, 15*time.Minute, 10)

	// then
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func Test_QuestionAPIClient_ReclaimStale_shouldReturnError_whenServerReturnsNon200(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)

	// when
	count, err := client.ReclaimStale(ctx, 15*time.Minute, 10)

	// then
	require.ErrorContains(t, err, "reclaim stale")
	assert.Equal(t, 0, count)
}
