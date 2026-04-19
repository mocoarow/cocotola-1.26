package domain

import "fmt"

const defaultMaxWorkbooks = 3

const maxAllowedWorkbooks = 100

// UserSetting holds per-user configuration such as resource limits.
type UserSetting struct {
	appUserID    AppUserID
	version      int
	maxWorkbooks int
}

// NewUserSetting creates a validated UserSetting.
func NewUserSetting(appUserID AppUserID, maxWorkbooks int) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// NewDefaultUserSetting creates a UserSetting with default values.
func NewDefaultUserSetting(appUserID AppUserID) (*UserSetting, error) {
	return NewUserSetting(appUserID, defaultMaxWorkbooks)
}

// ReconstructUserSetting reconstitutes a UserSetting from persistence.
func ReconstructUserSetting(appUserID AppUserID, maxWorkbooks int) (*UserSetting, error) {
	m := &UserSetting{
		appUserID:    appUserID,
		version:      0,
		maxWorkbooks: maxWorkbooks,
	}
	if err := m.validate(); err != nil {
		return nil, err
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
	return nil
}

// AppUserID returns the user ID.
func (s *UserSetting) AppUserID() AppUserID { return s.appUserID }

// MaxWorkbooks returns the maximum number of workbooks the user can create.
func (s *UserSetting) MaxWorkbooks() int { return s.maxWorkbooks }

// Version returns the persisted row version (0 = new, not yet saved).
func (s *UserSetting) Version() int { return s.version }

// SetVersion sets the persisted row version on a reconstituted aggregate.
func (s *UserSetting) SetVersion(version int) {
	s.version = version
}
