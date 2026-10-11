package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// UserSettingFinder is the lookup contract LoadOrInitUserSetting requires.
type UserSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
}

// LoadOrInitUserSetting returns the persisted UserSetting for userID, or a
// default one when no row exists. Use it everywhere so defaults cannot drift.
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
