package usersetting

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// userSettingFinderSaver is the persistence dependency of every user-setting update handler.
type userSettingFinderSaver interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
	Save(ctx context.Context, setting *domain.UserSetting) error
}

// UpdateLanguageHandler updates the authenticated user's preferred UI language.
type UpdateLanguageHandler struct {
	updateRunner
}

// NewUpdateLanguageHandler returns a new UpdateLanguageHandler.
func NewUpdateLanguageHandler(settingSaver userSettingFinderSaver) *UpdateLanguageHandler {
	return &UpdateLanguageHandler{newUpdateRunner("UpdateLanguageHandler", settingSaver)}
}

// UpdateLanguage handles PUT /auth/user-setting/language.
func (h *UpdateLanguageHandler) UpdateLanguage(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context) (changeFn, error) {
			var req api.UpdateUserLanguageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update language request: %w", err)
			}

			return func(s *domain.UserSetting) error { return s.ChangeLanguage(string(req.Language)) }, nil
		},
		invalidBindLog:   "invalid update language request",
		invalidChangeLog: "change language",
		invalidChangeMsg: "language is invalid",
	})
}
