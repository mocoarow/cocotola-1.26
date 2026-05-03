package domain

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/i18n"
)

const defaultMaxWorkbooks = 3

const maxAllowedWorkbooks = 100

const defaultLanguage = "en"

// DefaultLanguage returns the default ISO 639-1 language code applied to users
// without an explicit user-setting entry. Exposed so HTTP handlers can answer
// `language` queries without constructing a full UserSetting just for the
// default fallback.
func DefaultLanguage() string { return defaultLanguage }

// UserSetting holds per-user configuration such as resource limits.
type UserSetting struct {
	appUserID    AppUserID
	version      int
	maxWorkbooks int
	language     string
}

// NewUserSetting creates a validated UserSetting.
func NewUserSetting(appUserID AppUserID, maxWorkbooks int, language string) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
		language:     language,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("new user setting: %w", err)
	}
	return m, nil
}

// NewDefaultUserSetting creates a UserSetting with default values.
func NewDefaultUserSetting(appUserID AppUserID) (*UserSetting, error) {
	return NewUserSetting(appUserID, defaultMaxWorkbooks, defaultLanguage)
}

// ReconstructUserSetting reconstitutes a UserSetting from persistence.
func ReconstructUserSetting(appUserID AppUserID, maxWorkbooks int, language string) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
		language:     language,
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
	if !i18n.IsValidISO6391(s.language) {
		return fmt.Errorf("user setting language must be a lowercase ISO 639-1 code: %w", ErrInvalidArgument)
	}
	return nil
}

// AppUserID returns the user ID.
func (s *UserSetting) AppUserID() AppUserID { return s.appUserID }

// MaxWorkbooks returns the maximum number of workbooks the user can create.
func (s *UserSetting) MaxWorkbooks() int { return s.maxWorkbooks }

// Language returns the user's preferred language as an ISO 639-1 code.
func (s *UserSetting) Language() string { return s.language }

// ChangeLanguage updates the language. The new value must be a lowercase
// ISO 639-1 code (e.g. "ja", "en").
func (s *UserSetting) ChangeLanguage(language string) error {
	if !i18n.IsValidISO6391(language) {
		return fmt.Errorf("user setting language must be a lowercase ISO 639-1 code: %w", ErrInvalidArgument)
	}
	s.language = language
	return nil
}

// Version returns the persisted row version (0 = new, not yet saved).
func (s *UserSetting) Version() int { return s.version }

// SetVersion sets the persisted row version on a reconstituted aggregate.
func (s *UserSetting) SetVersion(version int) {
	s.version = version
}
