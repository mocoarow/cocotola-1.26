package domain

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"

	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
)

type idKind interface {
	label() string
}

type (
	organizationKind    struct{}
	appUserKind         struct{}
	groupKind           struct{}
	spaceKind           struct{}
	appUserProviderKind struct{}
)

func (organizationKind) label() string    { return "organization id" }
func (appUserKind) label() string         { return "app user id" }
func (groupKind) label() string           { return "group id" }
func (spaceKind) label() string           { return "space id" }
func (appUserProviderKind) label() string { return "app user provider id" }

// ID is a UUID-backed value-object identifier. Each kind K is a distinct type,
// so the compiler prevents mixing e.g. an OrganizationID with an AppUserID.
type ID[K idKind] struct {
	value uuid.UUID
}

// OrganizationID is the value-object identifier for an Organization aggregate.
type OrganizationID = ID[organizationKind]

// AppUserID is the value-object identifier for an AppUser aggregate.
type AppUserID = ID[appUserKind]

// GroupID is the value-object identifier for a Group aggregate.
type GroupID = ID[groupKind]

// SpaceID is the value-object identifier for a Space aggregate.
type SpaceID = ID[spaceKind]

// AppUserProviderID is the value-object identifier for an AppUserProvider entity.
type AppUserProviderID = ID[appUserProviderKind]

func kindLabel[K idKind]() string {
	var k K
	return k.label()
}

func newIDV7[K idKind]() (ID[K], error) {
	u, err := uuid.NewV7()
	if err != nil {
		return ID[K]{}, fmt.Errorf("generate %s: %w", kindLabel[K](), err)
	}
	return ID[K]{value: u}, nil
}

func parseID[K idKind](s string) (ID[K], error) {
	if s == "" {
		return ID[K]{}, fmt.Errorf("%s must not be empty: %w", kindLabel[K](), ErrInvalidArgument)
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return ID[K]{}, fmt.Errorf("parse %s: %w: %w", kindLabel[K](), err, ErrInvalidArgument)
	}
	return ID[K]{value: u}, nil
}

func mustParseID[K idKind](s string) ID[K] {
	id, err := parseID[K](s)
	if err != nil {
		panic(err)
	}
	return id
}

func newIDFromUUID[K idKind](u uuid.UUID) (ID[K], error) {
	if u == uuid.Nil {
		return ID[K]{}, fmt.Errorf("%s must not be zero: %w", kindLabel[K](), ErrInvalidArgument)
	}
	return ID[K]{value: u}, nil
}

// UUID returns the underlying uuid.UUID value.
func (id ID[K]) UUID() uuid.UUID { return id.value }

// String returns the canonical string representation.
func (id ID[K]) String() string { return id.value.String() }

// IsZero reports whether the ID has its zero value.
func (id ID[K]) IsZero() bool { return id.value == uuid.Nil }

// Equal reports whether two IDs are equal.
func (id ID[K]) Equal(other ID[K]) bool { return id.value == other.value }

// MarshalJSON encodes the ID as a JSON string.
func (id ID[K]) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.String() + `"`), nil
}

// UnmarshalJSON decodes a JSON string into the ID.
func (id *ID[K]) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	parsed, err := parseID[K](s)
	if err != nil {
		return fmt.Errorf("unmarshal %s: %w", kindLabel[K](), err)
	}

	*id = parsed

	return nil
}

// Value implements driver.Valuer so GORM/sql drivers can persist the ID as a string.
func (id ID[K]) Value() (driver.Value, error) {
	return id.String(), nil
}

// Scan implements sql.Scanner so drivers can load the ID back.
func (id *ID[K]) Scan(src any) error {
	if src == nil {
		*id = ID[K]{value: uuid.Nil}

		return nil
	}

	var s string

	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("scan %s: unsupported type %T: %w", kindLabel[K](), src, ErrInvalidArgument)
	}

	parsed, err := parseID[K](s)
	if err != nil {
		return fmt.Errorf("scan %s: %w", kindLabel[K](), err)
	}

	*id = parsed

	return nil
}

// NewOrganizationIDV7 generates a fresh UUIDv7-based OrganizationID.
func NewOrganizationIDV7() (OrganizationID, error) { return newIDV7[organizationKind]() }

// ParseOrganizationID parses a string into an OrganizationID.
func ParseOrganizationID(s string) (OrganizationID, error) { return parseID[organizationKind](s) }

// MustParseOrganizationID parses a string into an OrganizationID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseOrganizationID(s string) OrganizationID { return mustParseID[organizationKind](s) }

// NewOrganizationIDFromUUID wraps an existing uuid.UUID. Returns an error if the
// input is the zero UUID.
func NewOrganizationIDFromUUID(u uuid.UUID) (OrganizationID, error) {
	return newIDFromUUID[organizationKind](u)
}

// NewAppUserIDV7 generates a fresh UUIDv7-based AppUserID.
func NewAppUserIDV7() (AppUserID, error) { return newIDV7[appUserKind]() }

// ParseAppUserID parses a string into an AppUserID.
func ParseAppUserID(s string) (AppUserID, error) { return parseID[appUserKind](s) }

// MustParseAppUserID parses a string into an AppUserID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseAppUserID(s string) AppUserID { return mustParseID[appUserKind](s) }

// NewAppUserIDFromUUID wraps an existing uuid.UUID. Returns an error if the
// input is the zero UUID.
func NewAppUserIDFromUUID(u uuid.UUID) (AppUserID, error) { return newIDFromUUID[appUserKind](u) }

// NewGroupIDV7 generates a fresh UUIDv7-based GroupID.
func NewGroupIDV7() (GroupID, error) { return newIDV7[groupKind]() }

// ParseGroupID parses a string into a GroupID.
func ParseGroupID(s string) (GroupID, error) { return parseID[groupKind](s) }

// MustParseGroupID parses a string into a GroupID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseGroupID(s string) GroupID { return mustParseID[groupKind](s) }

// NewSpaceIDV7 generates a fresh UUIDv7-based SpaceID.
func NewSpaceIDV7() (SpaceID, error) { return newIDV7[spaceKind]() }

// ParseSpaceID parses a string into a SpaceID.
func ParseSpaceID(s string) (SpaceID, error) { return parseID[spaceKind](s) }

// MustParseSpaceID parses a string into a SpaceID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseSpaceID(s string) SpaceID { return mustParseID[spaceKind](s) }

// NewAppUserProviderIDV7 generates a fresh UUIDv7-based AppUserProviderID.
func NewAppUserProviderIDV7() (AppUserProviderID, error) { return newIDV7[appUserProviderKind]() }

// ParseAppUserProviderID parses a string into an AppUserProviderID.
func ParseAppUserProviderID(s string) (AppUserProviderID, error) {
	return parseID[appUserProviderKind](s)
}

// MustParseAppUserProviderID parses a string into an AppUserProviderID and panics on error.
// Intended for tests and hard-coded constants only.
func MustParseAppUserProviderID(s string) AppUserProviderID {
	return mustParseID[appUserProviderKind](s)
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
