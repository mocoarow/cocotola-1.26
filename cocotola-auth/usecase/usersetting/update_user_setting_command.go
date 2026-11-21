package usersetting

import (
	"context"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

type userSettingFinderSaver interface {
	FindByAppUserID(ctx context.Context, appUserID domain.AppUserID) (*domain.UserSetting, error)
	Save(ctx context.Context, setting *domain.UserSetting) error
}

// UpdateUserSettingCommand changes one field of a user's setting, creating the default setting first if none exists.
type UpdateUserSettingCommand struct {
	repo userSettingFinderSaver
}

// NewUpdateUserSettingCommand returns a new UpdateUserSettingCommand.
func NewUpdateUserSettingCommand(repo userSettingFinderSaver) *UpdateUserSettingCommand {
	return &UpdateUserSettingCommand{repo: repo}
}

// UpdateLanguage changes the user's preferred UI language.
func (c *UpdateUserSettingCommand) UpdateLanguage(ctx context.Context, input *usersettingservice.UpdateLanguageInput) error {
	return c.update(ctx, input.AppUserID, func(s *domain.UserSetting) error { return s.ChangeLanguage(input.Language) })
}

// UpdateDailyGoal changes the user's daily goal.
func (c *UpdateUserSettingCommand) UpdateDailyGoal(ctx context.Context, input *usersettingservice.UpdateDailyGoalInput) error {
	return c.update(ctx, input.AppUserID, func(s *domain.UserSetting) error { return s.ChangeDailyGoal(input.DailyGoal) })
}

// UpdateTimezone changes the user's IANA timezone.
func (c *UpdateUserSettingCommand) UpdateTimezone(ctx context.Context, input *usersettingservice.UpdateTimezoneInput) error {
	return c.update(ctx, input.AppUserID, func(s *domain.UserSetting) error { return s.ChangeTimezone(input.Timezone) })
}

func (c *UpdateUserSettingCommand) update(ctx context.Context, appUserID domain.AppUserID, change func(*domain.UserSetting) error) error {
	setting, err := domain.LoadOrInitUserSetting(ctx, c.repo, appUserID)
	if err != nil {
		return fmt.Errorf("load user setting %s: %w", appUserID, err)
	}
	if err := change(setting); err != nil {
		return fmt.Errorf("change user setting %s: %w: %w", appUserID, usersettingservice.ErrInvalidValue, err)
	}

	if err := c.repo.Save(ctx, setting); err != nil {
		return fmt.Errorf("save user setting %s: %w", appUserID, err)
	}

	return nil
}
