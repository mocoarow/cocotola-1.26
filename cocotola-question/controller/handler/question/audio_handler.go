package question

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

const (
	defaultPendingAudioLimit       = 50
	defaultReclaimStaleAudioLimit  = 50
	minReclaimStaleAudioSeconds    = 60
	maxReclaimStaleAudioSeconds    = 24 * 60 * 60
	defaultReclaimStaleAudioSecond = 15 * 60
)

const audioStatusKey = "status"

// AudioBatchUsecase is the use case interface required by AudioHandler.
type AudioBatchUsecase interface {
	ListPendingAudio(ctx context.Context, input *questionservice.ListPendingAudioInput) (*questionservice.ListPendingAudioOutput, error)
	ClaimAudio(ctx context.Context, input *questionservice.ClaimAudioInput) error
	CompleteAudio(ctx context.Context, input *questionservice.CompleteAudioInput) error
	FailAudio(ctx context.Context, input *questionservice.FailAudioInput) error
	ReclaimStaleAudio(ctx context.Context, input *questionservice.ReclaimStaleAudioInput) (*questionservice.ReclaimStaleAudioOutput, error)
}

// AudioHandler exposes internal endpoints used by cocotola-audio-generator.
type AudioHandler struct {
	usecase AudioBatchUsecase
	logger  *slog.Logger
}

// NewAudioHandler returns a new AudioHandler.
func NewAudioHandler(usecase AudioBatchUsecase) *AudioHandler {
	return &AudioHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "AudioHandler")),
	}
}

type pendingAudioResponse struct {
	Items []pendingAudioItemDTO `json:"items"`
}

type pendingAudioItemDTO struct {
	WorkbookID  string `json:"workbookId"`
	QuestionID  string `json:"questionId"`
	SourceText  string `json:"sourceText"`
	SourceLang  string `json:"sourceLang"`
	TargetText  string `json:"targetText"`
	TargetLang  string `json:"targetLang"`
	InputHash   string `json:"inputHash"`
	FailedTries int    `json:"failedTries"`
}

// ListPendingAudio handles GET /api/v1/internal/audio/questions/pending?limit=N.
func (h *AudioHandler) ListPendingAudio(c *gin.Context) {
	ctx := c.Request.Context()

	limit := defaultPendingAudioLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.logger.WarnContext(ctx, "invalid limit", slog.String("limit", raw))
			c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "limit must be a positive integer"))
			return
		}
		limit = parsed
	}

	input, err := questionservice.NewListPendingAudioInput(limit)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "list pending audio input", err)
		return
	}
	output, err := h.usecase.ListPendingAudio(ctx, input)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "list pending audio", err)
		return
	}
	resp := pendingAudioResponse{Items: make([]pendingAudioItemDTO, 0, len(output.Items))}
	for _, it := range output.Items {
		resp.Items = append(resp.Items, pendingAudioItemDTO{
			WorkbookID:  it.WorkbookID,
			QuestionID:  it.QuestionID,
			SourceText:  it.SourceText,
			SourceLang:  it.SourceLang,
			TargetText:  it.TargetText,
			TargetLang:  it.TargetLang,
			InputHash:   it.InputHash,
			FailedTries: it.FailedTries,
		})
	}
	c.JSON(http.StatusOK, resp)
}

type claimAudioRequest struct {
	InputHash string `json:"inputHash" binding:"required"`
}

// ClaimAudio handles POST /api/v1/internal/audio/questions/:workbookId/:questionId/claim.
func (h *AudioHandler) ClaimAudio(c *gin.Context) {
	ctx := c.Request.Context()
	workbookID, questionID, ok := h.requireIDs(ctx, c)
	if !ok {
		return
	}
	var req claimAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid claim audio request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "invalid request body"))
		return
	}
	input, err := questionservice.NewClaimAudioInput(workbookID, questionID, req.InputHash)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "claim audio input", err)
		return
	}
	if err := h.usecase.ClaimAudio(ctx, input); err != nil {
		handleAudioError(ctx, h.logger, c, "claim audio", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{audioStatusKey: "generating"})
}

type completeAudioRefRequest struct {
	Path        string  `json:"path" binding:"required"`
	DurationSec float64 `json:"durationSec"`
	SizeBytes   int64   `json:"sizeBytes"`
}

type completeAudioRequest struct {
	InputHash string                             `json:"inputHash" binding:"required"`
	Refs      map[string]completeAudioRefRequest `json:"refs" binding:"required"`
}

// CompleteAudio handles POST /api/v1/internal/audio/questions/:workbookId/:questionId/complete.
func (h *AudioHandler) CompleteAudio(c *gin.Context) {
	ctx := c.Request.Context()
	workbookID, questionID, ok := h.requireIDs(ctx, c)
	if !ok {
		return
	}
	var req completeAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid complete audio request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "invalid request body"))
		return
	}
	refs := make(map[string]questionservice.CompleteAudioRefInput, len(req.Refs))
	for k, v := range req.Refs {
		refs[k] = questionservice.CompleteAudioRefInput{
			Path:        v.Path,
			DurationSec: v.DurationSec,
			SizeBytes:   v.SizeBytes,
		}
	}
	input, err := questionservice.NewCompleteAudioInput(workbookID, questionID, req.InputHash, refs)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "complete audio input", err)
		return
	}
	if err := h.usecase.CompleteAudio(ctx, input); err != nil {
		handleAudioError(ctx, h.logger, c, "complete audio", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{audioStatusKey: "ready"})
}

type failAudioRequest struct {
	InputHash string `json:"inputHash" binding:"required"`
	Reason    string `json:"reason"`
}

// FailAudio handles POST /api/v1/internal/audio/questions/:workbookId/:questionId/fail.
func (h *AudioHandler) FailAudio(c *gin.Context) {
	ctx := c.Request.Context()
	workbookID, questionID, ok := h.requireIDs(ctx, c)
	if !ok {
		return
	}
	var req failAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid fail audio request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "invalid request body"))
		return
	}
	input, err := questionservice.NewFailAudioInput(workbookID, questionID, req.InputHash, req.Reason)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "fail audio input", err)
		return
	}
	if err := h.usecase.FailAudio(ctx, input); err != nil {
		handleAudioError(ctx, h.logger, c, "fail audio", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{audioStatusKey: "failed"})
}

func (h *AudioHandler) requireIDs(ctx context.Context, c *gin.Context) (string, string, bool) {
	workbookID := c.Param("workbookId")
	if workbookID == "" {
		h.logger.WarnContext(ctx, "missing workbook ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "workbook ID is required"))
		return "", "", false
	}
	questionID := c.Param("questionId")
	if questionID == "" {
		h.logger.WarnContext(ctx, "missing question ID")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "question ID is required"))
		return "", "", false
	}
	return workbookID, questionID, true
}

type reclaimStaleAudioRequest struct {
	StaleAfterSec int `json:"staleAfterSec"`
	Limit         int `json:"limit"`
}

type reclaimStaleAudioResponse struct {
	Reclaimed int `json:"reclaimed"`
}

// ReclaimStaleAudio handles POST /api/v1/internal/audio/questions/reclaim-stale.
// The batch invokes this once per run to transition any "generating" entries
// stuck longer than staleAfterSec back to "pending" (e.g. a previous worker
// crashed before completing or failing them).
func (h *AudioHandler) ReclaimStaleAudio(c *gin.Context) {
	ctx := c.Request.Context()
	var req reclaimStaleAudioRequest
	// Empty body is allowed: defaults are applied below.
	_ = c.ShouldBindJSON(&req)

	staleSec := req.StaleAfterSec
	if staleSec <= 0 {
		staleSec = defaultReclaimStaleAudioSecond
	}
	if staleSec < minReclaimStaleAudioSeconds || staleSec > maxReclaimStaleAudioSeconds {
		h.logger.WarnContext(ctx, "invalid staleAfterSec", slog.Int("staleAfterSec", staleSec))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "staleAfterSec out of range"))
		return
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultReclaimStaleAudioLimit
	}

	input, err := questionservice.NewReclaimStaleAudioInput(time.Duration(staleSec)*time.Second, limit)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "reclaim stale audio input", err)
		return
	}
	output, err := h.usecase.ReclaimStaleAudio(ctx, input)
	if err != nil {
		handleAudioError(ctx, h.logger, c, "reclaim stale audio", err)
		return
	}
	c.JSON(http.StatusOK, reclaimStaleAudioResponse{Reclaimed: output.Reclaimed})
}

// InitInternalAudioRouter mounts the audio batch endpoints under the internal
// router (protected by X-Service-Api-Key).
func InitInternalAudioRouter(handler *AudioHandler, parentRouterGroup gin.IRouter) {
	g := parentRouterGroup.Group("audio/questions")
	g.GET("pending", handler.ListPendingAudio)
	g.POST("reclaim-stale", handler.ReclaimStaleAudio)
	g.POST("/:workbookId/:questionId/claim", handler.ClaimAudio)
	g.POST("/:workbookId/:questionId/complete", handler.CompleteAudio)
	g.POST("/:workbookId/:questionId/fail", handler.FailAudio)
}

func handleAudioError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	if errors.Is(err, domain.ErrInvalidArgument) {
		logger.WarnContext(ctx, "invalid argument", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", http.StatusText(http.StatusBadRequest)))
		return
	}
	if errors.Is(err, domain.ErrQuestionNotFound) {
		logger.WarnContext(ctx, "question not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("question_not_found", "question not found"))
		return
	}
	if errors.Is(err, domain.ErrAudioNotPending) {
		logger.InfoContext(ctx, "audio not pending", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("audio_not_pending", "audio is not pending"))
		return
	}
	if errors.Is(err, domain.ErrAudioNotGenerating) {
		logger.InfoContext(ctx, "audio not generating", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("audio_not_generating", "audio is not generating"))
		return
	}
	if errors.Is(err, domain.ErrAudioInputHashMismatch) {
		logger.InfoContext(ctx, "audio input hash mismatch", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("audio_input_hash_mismatch", "audio input hash does not match"))
		return
	}
	if errors.Is(err, domain.ErrAudioConcurrentModification) {
		logger.InfoContext(ctx, "audio concurrent modification", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("audio_concurrent_modification", "audio was modified concurrently"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
