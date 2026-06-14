package usersetting

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// userSettingFinderSaver is the persistence dependency shared by every
// user-setting update handler in this package. Named explicitly with both
// roles so the contract is visible at the call site: every handler that
// embeds updateRunner needs to find the existing row (for the read-modify-
// write loop) and save the mutated value back.
type userSettingFinderSaver interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
	Save(ctx context.Context, setting *domain.UserSetting) error
}

// UpdateLanguageHandler exposes the authenticated user's preferred UI
// language as a mutable resource. The struct owns the persistence
// dependency and a slog logger tagged with its handler name; the actual
// bind / load / validate / save flow is delegated to updateRunner so this
// type only declares the field-specific request type and Change call.
type UpdateLanguageHandler struct {
	updateRunner
}

// NewUpdateLanguageHandler returns a new UpdateLanguageHandler.
func NewUpdateLanguageHandler(settingSaver userSettingFinderSaver) *UpdateLanguageHandler {
	return &UpdateLanguageHandler{newUpdateRunner("UpdateLanguageHandler", settingSaver)}
}

// UpdateLanguage handles PUT /auth/user-setting/language. The
// authenticated user's preferred language is replaced with the value in
// the request body. If no UserSetting row exists for the user, a default
// one is created with the requested language applied. Responses: 204 on
// success, 400 for an invalid request body or unsupported language code,
// 401 if the caller is not authenticated, 409 on a concurrent
// modification, 500 on persistence failure.
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
