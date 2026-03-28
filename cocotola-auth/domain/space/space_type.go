// Package space contains the space aggregate of the cocotola-auth domain.
package space

import "errors"

// Type represents the type of a space.
type Type struct {
	value string
}

const (
	spaceTypePublic  = "public"
	spaceTypePrivate = "private"
)

// TypePublic returns the public space type.
func TypePublic() Type { return Type{value: spaceTypePublic} }

// TypePrivate returns the private space type.
func TypePrivate() Type { return Type{value: spaceTypePrivate} }

// NewType creates a validated SpaceType from a string.
func NewType(value string) (Type, error) {
	switch value {
	case spaceTypePublic:
		return TypePublic(), nil
	case spaceTypePrivate:
		return TypePrivate(), nil
	default:
		return Type{}, errors.New("invalid space type: must be 'public' or 'private'")
	}
}

// Value returns the string representation.
func (t Type) Value() string { return t.value }

// IsPublic returns true if the space type is public.
func (t Type) IsPublic() bool { return t.value == spaceTypePublic }

// IsPrivate returns true if the space type is private.
func (t Type) IsPrivate() bool { return t.value == spaceTypePrivate }
