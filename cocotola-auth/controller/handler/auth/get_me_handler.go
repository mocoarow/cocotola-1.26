package auth

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

// GetMe handles GET /auth/me and returns the authenticated user's identity
// together with their preferred language. Missing user-setting rows fall back
// to the default language so that long-lived sessions without an explicit
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

	language, err := h.resolveLanguage(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "resolve user language", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
		return
	}

	c.JSON(http.StatusOK, api.GetMeResponse{
		UserID:           userID.UUID(),
		LoginID:          loginID,
		OrganizationName: organizationName,
		Language:         language,
	})
}

func (h *GetMeHandler) resolveLanguage(ctx context.Context, userID domain.AppUserID) (string, error) {
	setting, err := h.settingFinder.FindByAppUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserSettingNotFound) {
			return domain.DefaultLanguage(), nil
		}
		return "", fmt.Errorf("find user setting %s: %w", userID, err)
	}
	return setting.Language(), nil
}
