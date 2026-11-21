// Package usersetting provides service-layer types for user setting input validation.
package usersetting

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// ErrInvalidValue is returned when the requested value is rejected by the UserSetting aggregate.
var ErrInvalidValue = errors.New("invalid user setting value")

// UpdateLanguageInput holds the parameters for changing a user's preferred UI language.
type UpdateLanguageInput struct {
	AppUserID domain.AppUserID
	Language  string
}

// NewUpdateLanguageInput creates a validated UpdateLanguageInput.
func NewUpdateLanguageInput(appUserID domain.AppUserID, language string) (*UpdateLanguageInput, error) {
	if appUserID.IsZero() {
		return nil, fmt.Errorf("update language input app user id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return &UpdateLanguageInput{AppUserID: appUserID, Language: language}, nil
}

// UpdateDailyGoalInput holds the parameters for changing a user's daily goal.
type UpdateDailyGoalInput struct {
	AppUserID domain.AppUserID
	DailyGoal int
}

// NewUpdateDailyGoalInput creates a validated UpdateDailyGoalInput.
func NewUpdateDailyGoalInput(appUserID domain.AppUserID, dailyGoal int) (*UpdateDailyGoalInput, error) {
	if appUserID.IsZero() {
		return nil, fmt.Errorf("update daily goal input app user id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return &UpdateDailyGoalInput{AppUserID: appUserID, DailyGoal: dailyGoal}, nil
}

// UpdateTimezoneInput holds the parameters for changing a user's IANA timezone.
type UpdateTimezoneInput struct {
	AppUserID domain.AppUserID
	Timezone  string
}

// NewUpdateTimezoneInput creates a validated UpdateTimezoneInput.
func NewUpdateTimezoneInput(appUserID domain.AppUserID, timezone string) (*UpdateTimezoneInput, error) {
	if appUserID.IsZero() {
		return nil, fmt.Errorf("update timezone input app user id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return &UpdateTimezoneInput{AppUserID: appUserID, Timezone: timezone}, nil
}
