package usersetting

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

// UpdateLanguageHandler updates the authenticated user's preferred UI language.
type UpdateLanguageHandler struct {
	updateRunner
}

// NewUpdateLanguageHandler returns a new UpdateLanguageHandler.
func NewUpdateLanguageHandler(updater userSettingUpdater) *UpdateLanguageHandler {
	return &UpdateLanguageHandler{newUpdateRunner("UpdateLanguageHandler", updater)}
}

// UpdateLanguage handles PUT /auth/user-setting/language.
func (h *UpdateLanguageHandler) UpdateLanguage(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context, appUserID domain.AppUserID) (applyFn, error) {
			var req api.UpdateUserLanguageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update language request: %w", err)
			}

			input, err := usersettingservice.NewUpdateLanguageInput(appUserID, string(req.Language))
			if err != nil {
				return nil, fmt.Errorf("new update language input: %w", err)
			}

			return func(ctx context.Context) error { return h.updater.UpdateLanguage(ctx, input) }, nil
		},
		invalidBindLog:   "invalid update language request",
		invalidChangeLog: "change language",
		invalidChangeMsg: "language is invalid",
	})
}
