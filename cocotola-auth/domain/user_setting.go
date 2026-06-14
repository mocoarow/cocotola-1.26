package domain

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/i18n"
)

const defaultMaxWorkbooks = 3

const maxAllowedWorkbooks = 100

const defaultLanguage = "en"

const defaultDailyGoal = 10

const minAllowedDailyGoal = 1

const maxAllowedDailyGoal = 500

const defaultTimezone = "Asia/Tokyo"

// DefaultLanguage returns the default ISO 639-1 language code applied to users
// without an explicit user-setting entry. Exposed so HTTP handlers can answer
// `language` queries without constructing a full UserSetting just for the
// default fallback.
func DefaultLanguage() string { return defaultLanguage }

// DefaultDailyGoal returns the default number of problems per day used as the
// fallback when no user-setting entry exists.
func DefaultDailyGoal() int { return defaultDailyGoal }

// DefaultTimezone returns the default IANA timezone name applied to users
// without an explicit user-setting entry.
func DefaultTimezone() string { return defaultTimezone }

// UserSetting holds per-user configuration such as resource limits.
type UserSetting struct {
	appUserID    AppUserID
	version      int
	maxWorkbooks int
	language     string
	dailyGoal    int
	timezone     string
}

// NewUserSetting creates a validated UserSetting.
func NewUserSetting(appUserID AppUserID, maxWorkbooks int, language string, dailyGoal int, timezone string) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
		language:     language,
		dailyGoal:    dailyGoal,
		timezone:     timezone,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("new user setting: %w", err)
	}
	return m, nil
}

// NewDefaultUserSetting creates a UserSetting with default values.
func NewDefaultUserSetting(appUserID AppUserID) (*UserSetting, error) {
	return NewUserSetting(appUserID, defaultMaxWorkbooks, defaultLanguage, defaultDailyGoal, defaultTimezone)
}

// ReconstructUserSetting reconstitutes a UserSetting from persistence.
func ReconstructUserSetting(appUserID AppUserID, maxWorkbooks int, language string, dailyGoal int, timezone string) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
		language:     language,
		dailyGoal:    dailyGoal,
		timezone:     timezone,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("reconstruct user setting: %w", err)
	}
	return m, nil
}

func (s *UserSetting) validate() error {
	if s.appUserID.IsZero() {
		return fmt.Errorf("user setting app user id must not be zero: %w", ErrInvalidArgument)
	}
	if s.maxWorkbooks <= 0 {
		return fmt.Errorf("user setting max workbooks must be positive: %w", ErrInvalidArgument)
	}
	if s.maxWorkbooks > maxAllowedWorkbooks {
		return fmt.Errorf("user setting max workbooks exceeds limit %d: %w", maxAllowedWorkbooks, ErrInvalidArgument)
	}
	if err := validateLanguageValue(s.language); err != nil {
		return err
	}
	if err := validateDailyGoalValue(s.dailyGoal); err != nil {
		return err
	}
	if err := validateTimezoneValue(s.timezone); err != nil {
		return err
	}
	return nil
}

// validateLanguageValue checks the per-field rules for the language code so
// validate() and ChangeLanguage share the same gate.
func validateLanguageValue(v string) error {
	if !i18n.IsValidISO6391(v) {
		return fmt.Errorf("user setting language must be a lowercase ISO 639-1 code: %w", ErrInvalidArgument)
	}
	return nil
}

// validateDailyGoalValue checks the per-field rules for the daily problem
// goal so validate() and ChangeDailyGoal share the same gate.
func validateDailyGoalValue(v int) error {
	if v < minAllowedDailyGoal {
		return fmt.Errorf("user setting daily goal must be at least %d: %w", minAllowedDailyGoal, ErrInvalidArgument)
	}
	if v > maxAllowedDailyGoal {
		return fmt.Errorf("user setting daily goal exceeds limit %d: %w", maxAllowedDailyGoal, ErrInvalidArgument)
	}
	return nil
}

// validateTimezoneValue checks the per-field rules for the IANA timezone
// name so validate() and ChangeTimezone share the same gate.
func validateTimezoneValue(v string) error {
	if v == "" {
		return fmt.Errorf("user setting timezone must not be empty: %w", ErrInvalidArgument)
	}
	if _, err := time.LoadLocation(v); err != nil {
		return fmt.Errorf("user setting timezone must be a valid IANA name: %w", ErrInvalidArgument)
	}
	return nil
}

// AppUserID returns the user ID.
func (s *UserSetting) AppUserID() AppUserID { return s.appUserID }

// MaxWorkbooks returns the maximum number of workbooks the user can create.
func (s *UserSetting) MaxWorkbooks() int { return s.maxWorkbooks }

// Language returns the user's preferred language as an ISO 639-1 code.
func (s *UserSetting) Language() string { return s.language }

// DailyGoal returns the daily problem count goal used by the dashboard.
func (s *UserSetting) DailyGoal() int { return s.dailyGoal }

// Timezone returns the IANA timezone name used to bucket daily study stats.
func (s *UserSetting) Timezone() string { return s.timezone }

// ChangeLanguage updates the language. The new value must be a lowercase
// ISO 639-1 code (e.g. "ja", "en"). Validates first, mutates after, so a
// rejected value leaves the receiver unchanged.
func (s *UserSetting) ChangeLanguage(language string) error {
	if err := validateLanguageValue(language); err != nil {
		return err
	}
	s.language = language
	return nil
}

// ChangeDailyGoal updates the daily problem count goal. The new value must be
// within [minAllowedDailyGoal, maxAllowedDailyGoal]. Validates first, mutates
// after, so a rejected value leaves the receiver unchanged.
func (s *UserSetting) ChangeDailyGoal(dailyGoal int) error {
	if err := validateDailyGoalValue(dailyGoal); err != nil {
		return err
	}
	s.dailyGoal = dailyGoal
	return nil
}

// ChangeTimezone updates the timezone. The new value must be a non-empty
// IANA timezone name resolvable via time.LoadLocation. Validates first,
// mutates after, so a rejected value leaves the receiver unchanged.
func (s *UserSetting) ChangeTimezone(timezone string) error {
	if err := validateTimezoneValue(timezone); err != nil {
		return err
	}
	s.timezone = timezone
	return nil
}

// Version returns the persisted row version (0 = new, not yet saved).
func (s *UserSetting) Version() int { return s.version }

// SetVersion sets the persisted row version on a reconstituted aggregate.
func (s *UserSetting) SetVersion(version int) {
	s.version = version
}
