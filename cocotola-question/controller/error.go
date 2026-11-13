package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// NewErrorResponse creates an ErrorResponse with the given error code and message.
func NewErrorResponse(code string, message string) *api.ErrorResponse {
	return &api.ErrorResponse{
		Code:    code,
		Message: message,
	}
}

type errorMapping struct {
	target   error
	status   int
	code     string
	message  string
	logLevel slog.Level
}

func errorMappings() []errorMapping {
	return []errorMapping{
		{target: domain.ErrInvalidArgument, status: http.StatusBadRequest, code: "invalid_request", message: http.StatusText(http.StatusBadRequest), logLevel: slog.LevelWarn},
		{target: domain.ErrForbidden, status: http.StatusForbidden, code: "forbidden", message: http.StatusText(http.StatusForbidden), logLevel: slog.LevelWarn},
		{target: domain.ErrWorkbookNotFound, status: http.StatusNotFound, code: "workbook_not_found", message: "workbook not found", logLevel: slog.LevelWarn},
		{target: domain.ErrQuestionNotFound, status: http.StatusNotFound, code: "question_not_found", message: "question not found", logLevel: slog.LevelWarn},
		{target: domain.ErrStudyRecordNotFound, status: http.StatusNotFound, code: "study_record_not_found", message: "study record not found", logLevel: slog.LevelWarn},
		{target: domain.ErrReferenceNotFound, status: http.StatusNotFound, code: "reference_not_found", message: "workbook reference not found", logLevel: slog.LevelWarn},
		{target: domain.ErrOwnedWorkbookListNotFound, status: http.StatusNotFound, code: "owned_workbook_list_not_found", message: "owned workbook list not found", logLevel: slog.LevelWarn},
		{target: domain.ErrActiveQuestionListNotFound, status: http.StatusNotFound, code: "active_question_list_not_found", message: "active question list not found", logLevel: slog.LevelWarn},
		{target: domain.ErrOwnedWorkbookLimitReached, status: http.StatusConflict, code: "owned_workbook_limit_reached", message: "owned workbook limit reached", logLevel: slog.LevelWarn},
		{target: domain.ErrDuplicateReference, status: http.StatusConflict, code: "duplicate_reference", message: "workbook reference already exists", logLevel: slog.LevelWarn},
		{target: libversioned.ErrConcurrentModification, status: http.StatusConflict, code: "conflict", message: "resource was modified concurrently", logLevel: slog.LevelWarn},
		{target: domain.ErrAudioNotPending, status: http.StatusConflict, code: "audio_not_pending", message: "audio is not pending", logLevel: slog.LevelInfo},
		{target: domain.ErrAudioNotGenerating, status: http.StatusConflict, code: "audio_not_generating", message: "audio is not generating", logLevel: slog.LevelInfo},
		{target: domain.ErrAudioInputHashMismatch, status: http.StatusConflict, code: "audio_input_hash_mismatch", message: "audio input hash does not match", logLevel: slog.LevelInfo},
		{target: domain.ErrAudioConcurrentModification, status: http.StatusConflict, code: "audio_concurrent_modification", message: "audio was modified concurrently", logLevel: slog.LevelInfo},
	}
}

// HandleError writes the error response for err returned from an input constructor or a usecase.
func HandleError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	for _, m := range errorMappings() {
		if errors.Is(err, m.target) {
			logger.Log(ctx, m.logLevel, action, slog.String("error_code", m.code), slog.Any("error", err))
			c.JSON(m.status, NewErrorResponse(m.code, m.message))
			return
		}
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
