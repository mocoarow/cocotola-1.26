package question_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

const audioInternalBase = "/api/v1/internal/audio/questions"

// --- ListPendingAudio ---

func Test_AudioHandler_ListPendingAudio_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ListPendingAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.ListPendingAudioInput) bool {
		return in.Limit == 25
	})).Return(&questionservice.ListPendingAudioOutput{Items: nil}, nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, audioInternalBase+"/pending?limit=25", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_AudioHandler_ListPendingAudio_shouldReturn400_whenLimitIsNotInteger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, audioInternalBase+"/pending?limit=abc", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ListPendingAudio_shouldReturn400_whenLimitIsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, audioInternalBase+"/pending?limit=0", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

// --- ClaimAudio ---

func Test_AudioHandler_ClaimAudio_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.ClaimAudioInput) bool {
		return in.WorkbookID == fixtureWorkbookID && in.QuestionID == fixtureQuestionID && in.InputHash == fixtureAudioInputHash
	})).Return(nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_AudioHandler_ClaimAudio_shouldReturn400_whenInputHashMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(`{}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ClaimAudio_shouldReturn400_whenInputHashIsNotHex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when: 64 chars but contains 'z' (invalid hex)
	body := `{"inputHash":"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ClaimAudio_shouldReturn400_whenInputHashIsWrongLength(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when: 32 hex chars (md5 length, not sha256)
	body := `{"inputHash":"a1b2c3d4e5f60718293a4b5c6d7e8f90"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ClaimAudio_shouldReturn404_whenQuestionNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.Anything).Return(domain.ErrQuestionNotFound).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "question_not_found")
}

func Test_AudioHandler_ClaimAudio_shouldReturn409_whenAudioNotPending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.Anything).Return(domain.ErrAudioNotPending).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "audio_not_pending")
}

func Test_AudioHandler_ClaimAudio_shouldReturn409_whenInputHashMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.Anything).Return(domain.ErrAudioInputHashMismatch).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "audio_input_hash_mismatch")
}

func Test_AudioHandler_ClaimAudio_shouldReturn409_whenConcurrentModification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.Anything).Return(domain.ErrAudioConcurrentModification).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "audio_concurrent_modification")
}

func Test_AudioHandler_ClaimAudio_shouldReturn500_whenUsecaseFailsForUnexpectedReason(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ClaimAudio", mock.Anything, mock.Anything).Return(errors.New("network down")).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "internal_server_error")
}

// --- CompleteAudio ---

func Test_AudioHandler_CompleteAudio_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("CompleteAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.CompleteAudioInput) bool {
		return in.WorkbookID == fixtureWorkbookID && in.QuestionID == fixtureQuestionID &&
			in.InputHash == fixtureAudioInputHash &&
			len(in.Refs) == 1 && in.Refs["source"].Path == "audio/questions/q1/source.opus"
	})).Return(nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `","refs":{"source":{"path":"audio/questions/q1/source.opus","durationSec":1.5,"sizeBytes":1234}}}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completeURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_AudioHandler_CompleteAudio_shouldReturn409_whenAudioNotGenerating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("CompleteAudio", mock.Anything, mock.Anything).Return(domain.ErrAudioNotGenerating).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `","refs":{"source":{"path":"a/b.opus"}}}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completeURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "audio_not_generating")
}

// --- FailAudio ---

func Test_AudioHandler_FailAudio_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("FailAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.FailAudioInput) bool {
		return in.WorkbookID == fixtureWorkbookID && in.QuestionID == fixtureQuestionID &&
			in.InputHash == fixtureAudioInputHash && in.Reason == "tts: invalid voice"
	})).Return(nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"inputHash":"` + fixtureAudioInputHash + `","reason":"tts: invalid voice"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, failURL(), strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- ReclaimStaleAudio ---

func Test_AudioHandler_ReclaimStaleAudio_shouldUseDefaults_whenBodyIsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ReclaimStaleAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.ReclaimStaleAudioInput) bool {
		// default StaleAfter = 15 min, default Limit = 50
		return in.StaleAfter.Minutes() == 15 && in.Limit == 50
	})).Return(&questionservice.ReclaimStaleAudioOutput{Reclaimed: 0}, nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, audioInternalBase+"/reclaim-stale", bytes.NewReader(nil))
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_AudioHandler_ReclaimStaleAudio_shouldReturn400_whenStaleAfterSecBelowMin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: 10s is below the 60s minimum
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"staleAfterSec":10}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, audioInternalBase+"/reclaim-stale", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ReclaimStaleAudio_shouldReturn400_whenStaleAfterSecAboveMax(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: 1 day + 1 second is above the 24h ceiling
	usecase := NewMockAudioBatchUsecase(t)
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"staleAfterSec":86401}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, audioInternalBase+"/reclaim-stale", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorCode(t, readBytes(t, w.Body), "invalid_request")
}

func Test_AudioHandler_ReclaimStaleAudio_shouldReturn200WithReclaimedCount_whenSuccessful(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockAudioBatchUsecase(t)
	usecase.On("ReclaimStaleAudio", mock.Anything, mock.MatchedBy(func(in *questionservice.ReclaimStaleAudioInput) bool {
		return in.StaleAfter.Seconds() == 600 && in.Limit == 25
	})).Return(&questionservice.ReclaimStaleAudioOutput{Reclaimed: 7}, nil).Once()
	r := initInternalAudioRouter(ctx, t, usecase)
	w := httptest.NewRecorder()

	// when
	body := `{"staleAfterSec":600,"limit":25}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, audioInternalBase+"/reclaim-stale", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	jsonObj := parseJSON(t, readBytes(t, w.Body))
	expr := parseExpr(t, "$.reclaimed")
	got := expr.Get(jsonObj)
	require.Len(t, got, 1)
	assert.EqualValues(t, 7, got[0])
}

// --- helpers ---

func claimURL() string {
	return audioInternalBase + "/" + fixtureWorkbookID + "/" + fixtureQuestionID + "/claim"
}

func completeURL() string {
	return audioInternalBase + "/" + fixtureWorkbookID + "/" + fixtureQuestionID + "/complete"
}

func failURL() string {
	return audioInternalBase + "/" + fixtureWorkbookID + "/" + fixtureQuestionID + "/fail"
}
