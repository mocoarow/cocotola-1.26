package auth

import (
	"context"
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

type userSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
}

// GetMeHandler handles the GET /auth/me endpoint.
type GetMeHandler struct {
	settingFinder userSettingFinder
	logger        *slog.Logger
}

// NewGetMeHandler returns a new GetMeHandler.
func NewGetMeHandler(settingFinder userSettingFinder) *GetMeHandler {
	return &GetMeHandler{
		settingFinder: settingFinder,
		logger:        slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetMeHandler")),
	}
}

// userPreferences captures the user-setting fields exposed to /auth/me.
type userPreferences struct {
	language  string
	dailyGoal int
	timezone  string
}

// GetMe handles GET /auth/me and returns the authenticated user's identity
// together with their user-setting preferences. Missing user-setting rows
// fall back to default values so long-lived sessions without an explicit
// preference keep working.
func (h *GetMeHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := handler.GetAppUserIDFromContext(c)
	if !ok {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	loginID := c.GetString(controller.ContextFieldLoginID{})
	if loginID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing login ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationName := c.GetString(controller.ContextFieldOrganizationName{})

	prefs, err := h.resolvePreferences(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "resolve user preferences", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.GetMeResponse{
		UserID:           userID.UUID(),
		LoginID:          loginID,
		OrganizationName: organizationName,
		Language:         prefs.language,
		DailyGoal:        prefs.dailyGoal,
		Timezone:         prefs.timezone,
	})
}

func (h *GetMeHandler) resolvePreferences(ctx context.Context, userID domain.AppUserID) (userPreferences, error) {
	setting, err := handler.LoadOrInitUserSetting(ctx, h.settingFinder, userID)
	if err != nil {
		return userPreferences{}, fmt.Errorf("load or init user setting %s: %w", userID, err)
	}
	return userPreferences{
		language:  setting.Language(),
		dailyGoal: setting.DailyGoal(),
		timezone:  setting.Timezone(),
	}, nil
}
