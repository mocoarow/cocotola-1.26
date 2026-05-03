package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/google/uuid"

	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
)

// OrganizationID is the value-object identifier for an Organization aggregate.
// It wraps uuid.UUID so that the Go compiler prevents accidental mixing with
// other ID types (AppUserID, GroupID, ...) and raw strings.
type OrganizationID struct {
	value uuid.UUID
}

// NewOrganizationIDV7 generates a fresh UUIDv7-based OrganizationID.
func NewOrganizationIDV7() (OrganizationID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return OrganizationID{}, fmt.Errorf("generate organization id: %w", err)
	}
	return OrganizationID{value: u}, nil
}

// ParseOrganizationID parses a string into an OrganizationID.
func ParseOrganizationID(s string) (OrganizationID, error) {
	if s == "" {
		return OrganizationID{}, errors.New("organization id must not be empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return OrganizationID{}, fmt.Errorf("parse organization id: %w", err)
	}
	return OrganizationID{value: u}, nil
}

// MustParseOrganizationID parses a string into an OrganizationID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseOrganizationID(s string) OrganizationID {
	id, err := ParseOrganizationID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// NewOrganizationIDFromUUID wraps an existing uuid.UUID. Returns an error if the
// input is the zero UUID.
func NewOrganizationIDFromUUID(u uuid.UUID) (OrganizationID, error) {
	if u == uuid.Nil {
		return OrganizationID{}, errors.New("organization id must not be zero")
	}
	return OrganizationID{value: u}, nil
}

// UUID returns the underlying uuid.UUID value.
func (id OrganizationID) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id OrganizationID) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id OrganizationID) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id OrganizationID) Equal(other OrganizationID) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id OrganizationID) MarshalJSON() ([]byte, error) {
	s := `"` + id.String() + `"`

	return []byte(s), nil
}

// UnmarshalJSON decodes a JSON string into an OrganizationID.
func (id *OrganizationID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := ParseOrganizationID(s)
	if err != nil {
		return fmt.Errorf("unmarshal organization id: %w", err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id OrganizationID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *OrganizationID) Scan(src any) error {
	if src == nil {
		*id = OrganizationID{value: uuid.Nil}

		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := ParseOrganizationID(v)
		if err != nil {
			return fmt.Errorf("scan organization id: %w", err)
		}
		*id = parsed
		return nil
	case []byte:
		parsed, err := ParseOrganizationID(string(v))
		if err != nil {
			return fmt.Errorf("scan organization id: %w", err)
		}
		*id = parsed
		return nil
	default:
		return fmt.Errorf("scan organization id: unsupported type %T", src)
	}
}

// AppUserID is the value-object identifier for an AppUser aggregate.
type AppUserID struct {
	value uuid.UUID
}

// NewAppUserIDV7 generates a fresh UUIDv7-based AppUserID.
func NewAppUserIDV7() (AppUserID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return AppUserID{}, fmt.Errorf("generate app user id: %w", err)
	}
	return AppUserID{value: u}, nil
}

// ParseAppUserID parses a string into an AppUserID.
func ParseAppUserID(s string) (AppUserID, error) {
	if s == "" {
		return AppUserID{}, errors.New("app user id must not be empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return AppUserID{}, fmt.Errorf("parse app user id: %w", err)
	}
	return AppUserID{value: u}, nil
}

// MustParseAppUserID parses a string into an AppUserID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseAppUserID(s string) AppUserID {
	id, err := ParseAppUserID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// NewAppUserIDFromUUID wraps an existing uuid.UUID. Returns an error if the
// input is the zero UUID.
func NewAppUserIDFromUUID(u uuid.UUID) (AppUserID, error) {
	if u == uuid.Nil {
		return AppUserID{}, errors.New("app user id must not be zero")
	}
	return AppUserID{value: u}, nil
}

// UUID returns the underlying uuid.UUID value.
func (id AppUserID) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id AppUserID) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id AppUserID) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id AppUserID) Equal(other AppUserID) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id AppUserID) MarshalJSON() ([]byte, error) {
	s := `"` + id.String() + `"`

	return []byte(s), nil
}

// UnmarshalJSON decodes a JSON string into an AppUserID.
func (id *AppUserID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := ParseAppUserID(s)
	if err != nil {
		return fmt.Errorf("unmarshal app user id: %w", err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id AppUserID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *AppUserID) Scan(src any) error {
	if src == nil {
		*id = AppUserID{value: uuid.Nil}

		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := ParseAppUserID(v)
		if err != nil {
			return fmt.Errorf("scan app user id: %w", err)
		}
		*id = parsed
		return nil
	case []byte:
		parsed, err := ParseAppUserID(string(v))
		if err != nil {
			return fmt.Errorf("scan app user id: %w", err)
		}
		*id = parsed
		return nil
	default:
		return fmt.Errorf("scan app user id: unsupported type %T", src)
	}
}

// GroupID is the value-object identifier for a Group aggregate.
type GroupID struct {
	value uuid.UUID
}

// NewGroupIDV7 generates a fresh UUIDv7-based GroupID.
func NewGroupIDV7() (GroupID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return GroupID{}, fmt.Errorf("generate group id: %w", err)
	}

	return GroupID{value: u}, nil
}

// ParseGroupID parses a string into a GroupID.
func ParseGroupID(s string) (GroupID, error) {
	if s == "" {
		return GroupID{}, errors.New("group id must not be empty")
	}

	u, err := uuid.Parse(s)
	if err != nil {
		return GroupID{}, fmt.Errorf("parse group id: %w", err)
	}

	return GroupID{value: u}, nil
}

// MustParseGroupID parses a string into a GroupID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseGroupID(s string) GroupID {
	id, err := ParseGroupID(s)
	if err != nil {
		panic(err)
	}

	return id
}

// UUID returns the underlying uuid.UUID value.
func (id GroupID) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id GroupID) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id GroupID) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id GroupID) Equal(other GroupID) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id GroupID) MarshalJSON() ([]byte, error) {
	s := `"` + id.String() + `"`

	return []byte(s), nil
}

// UnmarshalJSON decodes a JSON string into a GroupID.
func (id *GroupID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := ParseGroupID(s)
	if err != nil {
		return fmt.Errorf("unmarshal group id: %w", err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id GroupID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *GroupID) Scan(src any) error {
	if src == nil {
		*id = GroupID{value: uuid.Nil}

		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := ParseGroupID(v)
		if err != nil {
			return fmt.Errorf("scan group id: %w", err)
		}

		*id = parsed

		return nil
	case []byte:
		parsed, err := ParseGroupID(string(v))
		if err != nil {
			return fmt.Errorf("scan group id: %w", err)
		}

		*id = parsed

		return nil
	default:
		return fmt.Errorf("scan group id: unsupported type %T", src)
	}
}

// SpaceID is the value-object identifier for a Space aggregate.
type SpaceID struct {
	value uuid.UUID
}

// NewSpaceIDV7 generates a fresh UUIDv7-based SpaceID.
func NewSpaceIDV7() (SpaceID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return SpaceID{}, fmt.Errorf("generate space id: %w", err)
	}

	return SpaceID{value: u}, nil
}

// ParseSpaceID parses a string into a SpaceID.
func ParseSpaceID(s string) (SpaceID, error) {
	if s == "" {
		return SpaceID{}, errors.New("space id must not be empty")
	}

	u, err := uuid.Parse(s)
	if err != nil {
		return SpaceID{}, fmt.Errorf("parse space id: %w", err)
	}

	return SpaceID{value: u}, nil
}

// MustParseSpaceID parses a string into a SpaceID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseSpaceID(s string) SpaceID {
	id, err := ParseSpaceID(s)
	if err != nil {
		panic(err)
	}

	return id
}

// UUID returns the underlying uuid.UUID value.
func (id SpaceID) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id SpaceID) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id SpaceID) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id SpaceID) Equal(other SpaceID) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id SpaceID) MarshalJSON() ([]byte, error) {
	s := `"` + id.String() + `"`

	return []byte(s), nil
}

// UnmarshalJSON decodes a JSON string into a SpaceID.
func (id *SpaceID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := ParseSpaceID(s)
	if err != nil {
		return fmt.Errorf("unmarshal space id: %w", err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id SpaceID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *SpaceID) Scan(src any) error {
	if src == nil {
		*id = SpaceID{value: uuid.Nil}

		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := ParseSpaceID(v)
		if err != nil {
			return fmt.Errorf("scan space id: %w", err)
		}

		*id = parsed

		return nil
	case []byte:
		parsed, err := ParseSpaceID(string(v))
		if err != nil {
			return fmt.Errorf("scan space id: %w", err)
		}

		*id = parsed

		return nil
	default:
		return fmt.Errorf("scan space id: unsupported type %T", src)
	}
}

// AppUserProviderID is the value-object identifier for an AppUserProvider entity.
type AppUserProviderID struct {
	value uuid.UUID
}

// NewAppUserProviderIDV7 generates a fresh UUIDv7-based AppUserProviderID.
func NewAppUserProviderIDV7() (AppUserProviderID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return AppUserProviderID{}, fmt.Errorf("generate app user provider id: %w", err)
	}
	return AppUserProviderID{value: u}, nil
}

// ParseAppUserProviderID parses a string into an AppUserProviderID.
func ParseAppUserProviderID(s string) (AppUserProviderID, error) {
	if s == "" {
		return AppUserProviderID{}, errors.New("app user provider id must not be empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return AppUserProviderID{}, fmt.Errorf("parse app user provider id: %w", err)
	}
	return AppUserProviderID{value: u}, nil
}

// MustParseAppUserProviderID parses a string into an AppUserProviderID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseAppUserProviderID(s string) AppUserProviderID {
	id, err := ParseAppUserProviderID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// UUID returns the underlying uuid.UUID value.
func (id AppUserProviderID) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id AppUserProviderID) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id AppUserProviderID) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id AppUserProviderID) Equal(other AppUserProviderID) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id AppUserProviderID) MarshalJSON() ([]byte, error) {
	s := `"` + id.String() + `"`

	return []byte(s), nil
}

// UnmarshalJSON decodes a JSON string into an AppUserProviderID.
func (id *AppUserProviderID) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := ParseAppUserProviderID(s)
	if err != nil {
		return fmt.Errorf("unmarshal app user provider id: %w", err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id AppUserProviderID) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *AppUserProviderID) Scan(src any) error {
	if src == nil {
		*id = AppUserProviderID{value: uuid.Nil}

		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := ParseAppUserProviderID(v)
		if err != nil {
			return fmt.Errorf("scan app user provider id: %w", err)
		}
		*id = parsed
		return nil
	case []byte:
		parsed, err := ParseAppUserProviderID(string(v))
		if err != nil {
			return fmt.Errorf("scan app user provider id: %w", err)
		}
		*id = parsed
		return nil
	default:
		return fmt.Errorf("scan app user provider id: unsupported type %T", src)
	}
}

// Well-known bootstrap IDs used by seed SQL. These must match the constants in
// init-{postgres,mysql} and supabase/migrations so that the system-admin rows
// inserted at DB-init time are referenced consistently from Go.
const (
	// SystemOrganizationIDString is the UUID of the bootstrap "system" organization.
	SystemOrganizationIDString = "00000000-0000-7000-8000-000000000001"
	// SystemAppUserIDString is the UUID of the bootstrap "__system_admin" user.
	SystemAppUserIDString = libdomain.SystemAppUserIDString
)

// SystemOrganizationID returns the bootstrap organization ID as a value object.
func SystemOrganizationID() OrganizationID {
	return MustParseOrganizationID(SystemOrganizationIDString)
}

// SystemAppUserID returns the bootstrap system-admin user ID as a value object.
func SystemAppUserID() AppUserID {
	return MustParseAppUserID(SystemAppUserIDString)
}
