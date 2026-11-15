package domain

import "fmt"

// SpaceType is the type of a cocotola-auth space.
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

// NewSpaceType creates a SpaceType from the value cocotola-auth returns.
func NewSpaceType(value string) (SpaceType, error) {
	switch value {
	case spaceTypePublic:
		return SpaceTypePublic(), nil
	case spaceTypePrivate:
		return SpaceTypePrivate(), nil
	default:
		return SpaceType{}, fmt.Errorf("space type %q: %w", value, ErrInvalidSpaceType)
	}
}

// IsPublic reports whether the space type is public.
func (s SpaceType) IsPublic() bool { return s.value == spaceTypePublic }

// IsPrivate reports whether the space type is private.
func (s SpaceType) IsPrivate() bool { return s.value == spaceTypePrivate }
