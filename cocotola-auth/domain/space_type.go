package domain

import "errors"

// SpaceType represents the type of a space.
type SpaceType struct {
	value string
}

const (
	spaceTypePublic  = "public"
	spaceTypePrivate = "private"
)

// SpaceTypePublic returns the public space type.
func SpaceTypePublic() SpaceType { return SpaceType{value: spaceTypePublic} }

// SpaceTypePrivate returns the private space type.
func SpaceTypePrivate() SpaceType { return SpaceType{value: spaceTypePrivate} }

// NewSpaceType creates a validated SpaceType from a string.
func NewSpaceType(value string) (SpaceType, error) {
	switch value {
	case spaceTypePublic:
		return SpaceTypePublic(), nil
	case spaceTypePrivate:
		return SpaceTypePrivate(), nil
	default:
		return SpaceType{}, errors.New("invalid space type: must be 'public' or 'private'")
	}
}

// Value returns the string representation.
func (t SpaceType) Value() string { return t.value }

// IsPublic returns true if the space type is public.
func (t SpaceType) IsPublic() bool { return t.value == spaceTypePublic }

// IsPrivate returns true if the space type is private.
func (t SpaceType) IsPrivate() bool { return t.value == spaceTypePrivate }
