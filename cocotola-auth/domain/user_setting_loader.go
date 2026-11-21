package domain

import (
	"context"
	"errors"
	"fmt"
)

// UserSettingFinder is the lookup contract LoadOrInitUserSetting requires.
type UserSettingFinder interface {
	FindByAppUserID(ctx context.Context, appUserID AppUserID) (*UserSetting, error)
}

// LoadOrInitUserSetting returns the persisted UserSetting for userID, or a
// default one when no row exists. Use it everywhere so defaults cannot drift.
func LoadOrInitUserSetting(ctx context.Context, finder UserSettingFinder, userID AppUserID) (*UserSetting, error) {
	setting, err := finder.FindByAppUserID(ctx, userID)
	if err == nil {
		return setting, nil
	}

	if !errors.Is(err, ErrUserSettingNotFound) {
		return nil, fmt.Errorf("find user setting %s: %w", userID, err)
	}

	defaultSetting, err := NewDefaultUserSetting(userID)
	if err != nil {
		return nil, fmt.Errorf("new default user setting %s: %w", userID, err)
	}

	return defaultSetting, nil
}
