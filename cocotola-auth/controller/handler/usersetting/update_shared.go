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
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// userSettingUpdater is the command dependency of every user-setting update handler.
type userSettingUpdater interface {
	UpdateLanguage(ctx context.Context, input *usersettingservice.UpdateLanguageInput) error
	UpdateDailyGoal(ctx context.Context, input *usersettingservice.UpdateDailyGoalInput) error
	UpdateTimezone(ctx context.Context, input *usersettingservice.UpdateTimezoneInput) error
}

// applyFn runs the bound update command.
type applyFn func(ctx context.Context) error

// bindFn parses the per-endpoint request payload into an applyFn.
type bindFn func(c *gin.Context, appUserID domain.AppUserID) (applyFn, error)

// updateConfig holds the per-handler pieces plugged into updateRunner.run.
type updateConfig struct {
	bindRequest      bindFn
	invalidBindLog   string
	invalidChangeLog string
	invalidChangeMsg string
}

// updateRunner is the shared bind / apply / respond scaffold embedded by every update handler.
type updateRunner struct {
	updater userSettingUpdater
	logger  *slog.Logger
}

// newUpdateRunner returns an updateRunner whose logger is named handlerName.
func newUpdateRunner(handlerName string, updater userSettingUpdater) updateRunner {
	return updateRunner{
		updater: updater,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, handlerName)),
	}
}

// run executes the bind / apply / respond flow.
func (r updateRunner) run(c *gin.Context, cfg updateConfig) {
	ctx := c.Request.Context()

	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		r.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))

		return
	}

	apply, err := cfg.bindRequest(c, userID)
	if err != nil {
		r.logger.WarnContext(ctx, cfg.invalidBindLog, slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))

		return
	}

	r.writeResult(ctx, c, cfg, apply(ctx))
}

// writeResult writes the HTTP response for the result of an update command.
func (r updateRunner) writeResult(ctx context.Context, c *gin.Context, cfg updateConfig, err error) {
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, usersettingservice.ErrInvalidValue):
		r.logger.WarnContext(ctx, cfg.invalidChangeLog, slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", cfg.invalidChangeMsg))
	case errors.Is(err, libversioned.ErrConcurrentModification):
		r.logger.WarnContext(ctx, "concurrent modification", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", "user setting was modified concurrently"))
	case errors.Is(err, domain.ErrUserSettingNotFound):
		r.logger.WarnContext(ctx, "user setting not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("user_setting_not_found", "user setting not found"))
	default:
		r.logger.ErrorContext(ctx, "update user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
	}
}
