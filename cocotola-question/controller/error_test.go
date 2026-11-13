package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func handleError(t *testing.T, err error) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	controller.HandleError(context.Background(), slog.New(slog.DiscardHandler), c, "test action", err)
	return w
}

func handleErrorLogLevel(t *testing.T, err error) string {
	t.Helper()

	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	controller.HandleError(context.Background(), logger, c, "test action", err)

	var record struct {
		Level string `json:"level"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	return record.Level
}

func decodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) api.ErrorResponse {
	t.Helper()

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func Test_HandleError_shouldWriteMappedResponse_whenErrorWrapsKnownSentinel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sentinel    error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{name: "invalid argument", sentinel: domain.ErrInvalidArgument, wantStatus: http.StatusBadRequest, wantCode: "invalid_request", wantMessage: "Bad Request"},
		{name: "forbidden", sentinel: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: "forbidden", wantMessage: "Forbidden"},
		{name: "workbook not found", sentinel: domain.ErrWorkbookNotFound, wantStatus: http.StatusNotFound, wantCode: "workbook_not_found", wantMessage: "workbook not found"},
		{name: "question not found", sentinel: domain.ErrQuestionNotFound, wantStatus: http.StatusNotFound, wantCode: "question_not_found", wantMessage: "question not found"},
		{name: "study record not found", sentinel: domain.ErrStudyRecordNotFound, wantStatus: http.StatusNotFound, wantCode: "study_record_not_found", wantMessage: "study record not found"},
		{name: "reference not found", sentinel: domain.ErrReferenceNotFound, wantStatus: http.StatusNotFound, wantCode: "reference_not_found", wantMessage: "workbook reference not found"},
		{name: "owned workbook list not found", sentinel: domain.ErrOwnedWorkbookListNotFound, wantStatus: http.StatusNotFound, wantCode: "owned_workbook_list_not_found", wantMessage: "owned workbook list not found"},
		{name: "active question list not found", sentinel: domain.ErrActiveQuestionListNotFound, wantStatus: http.StatusNotFound, wantCode: "active_question_list_not_found", wantMessage: "active question list not found"},
		{name: "owned workbook limit reached", sentinel: domain.ErrOwnedWorkbookLimitReached, wantStatus: http.StatusConflict, wantCode: "owned_workbook_limit_reached", wantMessage: "owned workbook limit reached"},
		{name: "duplicate reference", sentinel: domain.ErrDuplicateReference, wantStatus: http.StatusConflict, wantCode: "duplicate_reference", wantMessage: "workbook reference already exists"},
		{name: "concurrent modification", sentinel: libversioned.ErrConcurrentModification, wantStatus: http.StatusConflict, wantCode: "conflict", wantMessage: "resource was modified concurrently"},
		{name: "audio not pending", sentinel: domain.ErrAudioNotPending, wantStatus: http.StatusConflict, wantCode: "audio_not_pending", wantMessage: "audio is not pending"},
		{name: "audio not generating", sentinel: domain.ErrAudioNotGenerating, wantStatus: http.StatusConflict, wantCode: "audio_not_generating", wantMessage: "audio is not generating"},
		{name: "audio input hash mismatch", sentinel: domain.ErrAudioInputHashMismatch, wantStatus: http.StatusConflict, wantCode: "audio_input_hash_mismatch", wantMessage: "audio input hash does not match"},
		{name: "audio concurrent modification", sentinel: domain.ErrAudioConcurrentModification, wantStatus: http.StatusConflict, wantCode: "audio_concurrent_modification", wantMessage: "audio was modified concurrently"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			err := fmt.Errorf("usecase: %w", tt.sentinel)

			// when
			w := handleError(t, err)

			// then
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, api.ErrorResponse{Code: tt.wantCode, Message: tt.wantMessage}, decodeErrorResponse(t, w))
		})
	}
}

func Test_HandleError_shouldWrite500_whenErrorIsUnknown(t *testing.T) {
	t.Parallel()

	// given
	err := errors.New("database is down")

	// when
	w := handleError(t, err)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, api.ErrorResponse{Code: "internal_server_error", Message: "Internal Server Error"}, decodeErrorResponse(t, w))
}

func Test_HandleError_shouldLogAtMappedLevel_whenErrorWrapsKnownSentinel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sentinel  error
		wantLevel slog.Level
	}{
		{name: "invalid argument", sentinel: domain.ErrInvalidArgument, wantLevel: slog.LevelWarn},
		{name: "forbidden", sentinel: domain.ErrForbidden, wantLevel: slog.LevelWarn},
		{name: "workbook not found", sentinel: domain.ErrWorkbookNotFound, wantLevel: slog.LevelWarn},
		{name: "question not found", sentinel: domain.ErrQuestionNotFound, wantLevel: slog.LevelWarn},
		{name: "study record not found", sentinel: domain.ErrStudyRecordNotFound, wantLevel: slog.LevelWarn},
		{name: "reference not found", sentinel: domain.ErrReferenceNotFound, wantLevel: slog.LevelWarn},
		{name: "owned workbook list not found", sentinel: domain.ErrOwnedWorkbookListNotFound, wantLevel: slog.LevelWarn},
		{name: "active question list not found", sentinel: domain.ErrActiveQuestionListNotFound, wantLevel: slog.LevelWarn},
		{name: "owned workbook limit reached", sentinel: domain.ErrOwnedWorkbookLimitReached, wantLevel: slog.LevelWarn},
		{name: "duplicate reference", sentinel: domain.ErrDuplicateReference, wantLevel: slog.LevelWarn},
		{name: "concurrent modification", sentinel: libversioned.ErrConcurrentModification, wantLevel: slog.LevelWarn},
		{name: "audio not pending", sentinel: domain.ErrAudioNotPending, wantLevel: slog.LevelInfo},
		{name: "audio not generating", sentinel: domain.ErrAudioNotGenerating, wantLevel: slog.LevelInfo},
		{name: "audio input hash mismatch", sentinel: domain.ErrAudioInputHashMismatch, wantLevel: slog.LevelInfo},
		{name: "audio concurrent modification", sentinel: domain.ErrAudioConcurrentModification, wantLevel: slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			err := fmt.Errorf("usecase: %w", tt.sentinel)

			// when
			level := handleErrorLogLevel(t, err)

			// then
			assert.Equal(t, tt.wantLevel.String(), level)
		})
	}
}

func Test_HandleError_shouldLogAtErrorLevel_whenErrorIsUnknown(t *testing.T) {
	t.Parallel()

	// given
	err := errors.New("database is down")

	// when
	level := handleErrorLogLevel(t, err)

	// then
	assert.Equal(t, slog.LevelError.String(), level)
}
