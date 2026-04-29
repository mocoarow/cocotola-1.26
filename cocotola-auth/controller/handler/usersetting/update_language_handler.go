package usersetting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

type userSettingSaver interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
	Save(ctx context.Context, setting *domain.UserSetting) error
}

// UpdateLanguageHandler handles the PUT /auth/user-setting/language endpoint.
type UpdateLanguageHandler struct {
	settingSaver userSettingSaver
	logger       *slog.Logger
}

// NewUpdateLanguageHandler returns a new UpdateLanguageHandler.
func NewUpdateLanguageHandler(settingSaver userSettingSaver) *UpdateLanguageHandler {
	return &UpdateLanguageHandler{
		settingSaver: settingSaver,
		logger:       slog.Default().With(slog.String(liblogging.LoggerNameKey, "UpdateLanguageHandler")),
	}
}

// UpdateLanguage handles PUT /auth/user-setting/language. The authenticated
// user's preferred language is replaced with the value in the request body.
// If no UserSetting row exists for the user, a default one is created with
// the requested language applied.
func (h *UpdateLanguageHandler) UpdateLanguage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	var req api.UpdateUserLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid update language request", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "request body is invalid"))
		return
	}

	setting, err := h.loadOrInitSetting(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "load user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	if err := setting.ChangeLanguage(string(req.Language)); err != nil {
		h.logger.WarnContext(ctx, "change language", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "language is invalid"))
		return
	}

	if err := h.settingSaver.Save(ctx, setting); err != nil {
		if errors.Is(err, domain.ErrUserSettingConcurrentModification) {
			h.logger.WarnContext(ctx, "concurrent modification", slog.Any("error", err))
			c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", "user setting was modified concurrently"))
			return
		}
		h.logger.ErrorContext(ctx, "save user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UpdateLanguageHandler) loadOrInitSetting(ctx context.Context, userID domain.AppUserID) (*domain.UserSetting, error) {
	setting, err := h.settingSaver.FindByAppUserID(ctx, userID)
	if err == nil {
		return setting, nil
	}
	if !errors.Is(err, domain.ErrUserSettingNotFound) {
		return nil, fmt.Errorf("find user setting %s: %w", userID, err)
	}
	defaultSetting, err := domain.NewDefaultUserSetting(userID)
	if err != nil {
		return nil, fmt.Errorf("new default user setting %s: %w", userID, err)
	}
	return defaultSetting, nil
}
