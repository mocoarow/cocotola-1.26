// Package usersetting provides HTTP handlers for user setting operations.
package usersetting

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

type userSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
}

type findUserSettingResponse struct {
	MaxWorkbooks int `json:"maxWorkbooks"`
}

// FindUserSettingHandler handles the GET /auth/user-setting endpoint.
type FindUserSettingHandler struct {
	settingFinder userSettingFinder
	logger        *slog.Logger
}

// NewFindUserSettingHandler returns a new FindUserSettingHandler.
func NewFindUserSettingHandler(settingFinder userSettingFinder) *FindUserSettingHandler {
	return &FindUserSettingHandler{
		settingFinder: settingFinder,
		logger:        slog.Default().With(slog.String(liblogging.LoggerNameKey, "FindUserSettingHandler")),
	}
}

// FindUserSetting handles GET /auth/user-setting?user_id=<id>.
func (h *FindUserSettingHandler) FindUserSetting(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.WarnContext(ctx, "missing user_id query parameter")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "user_id query parameter is required"))
		return
	}

	appUserID, err := domain.ParseAppUserID(userIDStr)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", slog.String("user_id", userIDStr), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("bad_request", "invalid user_id"))
		return
	}

	setting, err := h.settingFinder.FindByAppUserID(ctx, appUserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserSettingNotFound) {
			// Return default values when no setting exists.
			defaultSetting, defErr := domain.NewDefaultUserSetting(appUserID)
			if defErr != nil {
				h.logger.ErrorContext(ctx, "create default user setting", slog.Any("error", defErr))
				c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
				return
			}
			c.JSON(http.StatusOK, findUserSettingResponse{
				MaxWorkbooks: defaultSetting.MaxWorkbooks(),
			})
			return
		}
		h.logger.ErrorContext(ctx, "find user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, findUserSettingResponse{
		MaxWorkbooks: setting.MaxWorkbooks(),
	})
}

// InitUserSettingRouter sets up the routes for user setting operations.
func InitUserSettingRouter(
	findHandler *FindUserSettingHandler,
	parentRouterGroup gin.IRouter,
	middleware ...gin.HandlerFunc,
) {
	settingGroup := parentRouterGroup.Group("user-setting")

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, findHandler.FindUserSetting)
	settingGroup.GET("", handlers...)
}
