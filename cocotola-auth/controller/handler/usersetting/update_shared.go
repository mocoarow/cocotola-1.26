package usersetting

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// changeFn applies an update to a UserSetting in place. It exists so the
// generic runUpdate pipeline can plug in any field-specific mutation
// (ChangeDailyGoal, ChangeTimezone, ...) without each handler having to
// repeat the bind / load / validate / save scaffolding.
type changeFn func(*domain.UserSetting) error

// bindFn parses and validates the per-endpoint request payload. The
// returned error is surfaced via the structured logger so a malformed
// payload still produces an actionable diagnostic — the legacy bool
// signaling discarded validator details (`required`, `oneof`, `max`,
// etc.) and reduced server logs to "invalid update <field> request"
// with no context.
type bindFn func(c *gin.Context) (changeFn, error)

// updateConfig collapses the per-handler pieces (request binder, mutator,
// invalid-mutation message) so updateRunner.run is the only place the HTTP
// response shape lives.
type updateConfig struct {
	bindRequest      bindFn
	invalidBindLog   string
	invalidChangeLog string
	invalidChangeMsg string
}

// updateRunner is the shared scaffold every user-setting update handler
// composes via struct embedding. It owns the userSettingFinderSaver
// dependency and the slog logger so each concrete handler file only
// declares the field-specific request type and Change* call.
type updateRunner struct {
	settingSaver userSettingFinderSaver
	logger       *slog.Logger
}

// newUpdateRunner returns an updateRunner whose logger is tagged with the
// supplied handler name (used as the slog logger-name key).
func newUpdateRunner(handlerName string, saver userSettingFinderSaver) updateRunner {
	return updateRunner{
		settingSaver: saver,
		logger:       slog.Default().With(slog.String(liblogging.LoggerNameKey, handlerName)),
	}
}

// run executes the canonical bind / load / mutate / save flow used by
// every user-setting update endpoint. Each handler supplies the differing
// pieces via updateConfig.
func (r updateRunner) run(c *gin.Context, cfg updateConfig) {
	ctx := c.Request.Context()

	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		r.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))

		return
	}

	change, err := cfg.bindRequest(c)
	if err != nil {
		r.logger.WarnContext(ctx, cfg.invalidBindLog, slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))

		return
	}

	setting, err := handler.LoadOrInitUserSetting(ctx, r.settingSaver, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "load user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))

		return
	}

	if err := change(setting); err != nil {
		r.logger.WarnContext(ctx, cfg.invalidChangeLog, slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", cfg.invalidChangeMsg))

		return
	}

	if writeSaveError(ctx, r.logger, c, r.settingSaver.Save(ctx, setting)) {
		return
	}

	c.Status(http.StatusNoContent)
}

// writeSaveError maps the gateway error returned by userSettingFinderSaver.Save
// onto the canonical HTTP response set used by every user-setting update
// handler. Returns true when an error response was written (handler should
// return immediately), false when no error to handle.
func writeSaveError(ctx context.Context, logger *slog.Logger, c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, libversioned.ErrConcurrentModification) {
		logger.WarnContext(ctx, "concurrent modification", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", "user setting was modified concurrently"))

		return true
	}

	if errors.Is(err, domain.ErrUserSettingNotFound) {
		logger.WarnContext(ctx, "user setting not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("user_setting_not_found", "user setting not found"))

		return true
	}

	logger.ErrorContext(ctx, "save user setting", slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))

	return true
}
