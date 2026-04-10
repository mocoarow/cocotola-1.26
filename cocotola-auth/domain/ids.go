package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	return json.Marshal(id.String())
}

// UnmarshalJSON decodes a JSON string into an OrganizationID.
func (id *OrganizationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal organization id: %w", err)
	}
	parsed, err := ParseOrganizationID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id OrganizationID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *OrganizationID) Scan(src any) error {
	if src == nil {
		*id = OrganizationID{}
		return nil
	}
	switch v := src.(type) {
	case string:
		parsed, err := ParseOrganizationID(v)
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	case []byte:
		parsed, err := ParseOrganizationID(string(v))
		if err != nil {
			return err
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
	return json.Marshal(id.String())
}

// UnmarshalJSON decodes a JSON string into an AppUserID.
func (id *AppUserID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal app user id: %w", err)
	}
	parsed, err := ParseAppUserID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id AppUserID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *AppUserID) Scan(src any) error {
	if src == nil {
		*id = AppUserID{}
		return nil
	}
	switch v := src.(type) {
	case string:
		parsed, err := ParseAppUserID(v)
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	case []byte:
		parsed, err := ParseAppUserID(string(v))
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	default:
		return fmt.Errorf("scan app user id: unsupported type %T", src)
	}
}

// Well-known bootstrap IDs used by seed SQL. These must match the constants in
// init-{postgres,mysql} and supabase/migrations so that the system-admin rows
// inserted at DB-init time are referenced consistently from Go.
const (
	// SystemOrganizationIDString is the UUID of the bootstrap "system" organization.
	SystemOrganizationIDString = "00000000-0000-7000-8000-000000000001"
	// SystemAppUserIDString is the UUID of the bootstrap "__system_admin" user.
	SystemAppUserIDString = "00000000-0000-7000-8000-000000000002"
)

// SystemOrganizationID returns the bootstrap organization ID as a value object.
func SystemOrganizationID() OrganizationID {
	return MustParseOrganizationID(SystemOrganizationIDString)
}

// SystemAppUserID returns the bootstrap system-admin user ID as a value object.
func SystemAppUserID() AppUserID {
	return MustParseAppUserID(SystemAppUserIDString)
}
