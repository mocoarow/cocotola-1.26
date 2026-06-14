package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// UserSettingFinder is the minimal lookup contract LoadOrInitUserSetting
// requires. The concrete *gateway.UserSettingRepository satisfies it, as
// does any local sub-package interface that exposes FindByAppUserID.
type UserSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
}

// LoadOrInitUserSetting returns the persisted UserSetting for userID, or a
// freshly constructed default when no row exists. Centralizing the
// "no row -> default object" behavior here keeps the contract identical
// across every endpoint that surfaces user-setting fields:
//
//   - GET  /auth/me                   (read-only)
//   - GET  /auth/user-setting         (read-only)
//   - PUT  /auth/user-setting/...     (read-modify-write)
//
// Without this helper, each call site would either re-implement the
// errors.Is(ErrUserSettingNotFound) branch (and risk drifting on default
// values) or short-circuit by reading domain.Default* constants and
// hand-assembling response structs — which is exactly the divergence the
// reviewers flagged.
func LoadOrInitUserSetting(ctx context.Context, finder UserSettingFinder, userID domain.AppUserID) (*domain.UserSetting, error) {
	setting, err := finder.FindByAppUserID(ctx, userID)
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
