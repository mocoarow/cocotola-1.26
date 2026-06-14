// Package usersetting provides HTTP handlers for user setting operations.
package usersetting

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

type userSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
}

type findUserSettingResponse struct {
	MaxWorkbooks int    `json:"maxWorkbooks"`
	Language     string `json:"language"`
	DailyGoal    int    `json:"dailyGoal"`
	Timezone     string `json:"timezone"`
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

	setting, err := handler.LoadOrInitUserSetting(ctx, h.settingFinder, appUserID)
	if err != nil {
		h.logger.ErrorContext(ctx, "load user setting", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, findUserSettingResponse{
		MaxWorkbooks: setting.MaxWorkbooks(),
		Language:     setting.Language(),
		DailyGoal:    setting.DailyGoal(),
		Timezone:     setting.Timezone(),
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

// WireExternalUserSettingHandlers constructs every external user-setting
// update handler from the supplied saver and mounts them under the parent
// router group. Single entry point so main.go (standalone deployment) and
// initialize/ (cocotola-app composite deployment) do not have to keep the
// per-endpoint constructor list synchronized; adding a new update endpoint
// only requires touching this function (plus its handler file and the
// router setup below).
func WireExternalUserSettingHandlers(
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	settingSaver userSettingFinderSaver,
	middleware ...gin.HandlerFunc,
) {
	settingGroup := parentRouterGroup.Group("user-setting")
	settingGroup.Use(authMiddleware)
	settingGroup.Use(middleware...)

	settingGroup.PUT("/language", NewUpdateLanguageHandler(settingSaver).UpdateLanguage)
	settingGroup.PUT("/daily-goal", NewUpdateDailyGoalHandler(settingSaver).UpdateDailyGoal)
	settingGroup.PUT("/timezone", NewUpdateTimezoneHandler(settingSaver).UpdateTimezone)
}
