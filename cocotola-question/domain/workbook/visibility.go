// Package workbook contains the workbook aggregate of the cocotola-question domain.
package workbook

import "errors"

// Visibility represents the visibility setting of a workbook.
type Visibility struct {
	value string
}

const (
	visibilityPrivate = "private"
	visibilityPublic  = "public"
)

// VisibilityPrivate returns the private visibility.
func VisibilityPrivate() Visibility { return Visibility{value: visibilityPrivate} }

// VisibilityPublic returns the public visibility.
func VisibilityPublic() Visibility { return Visibility{value: visibilityPublic} }

// NewVisibility creates a validated Visibility from a string.
func NewVisibility(value string) (Visibility, error) {
	switch value {
	case visibilityPrivate:
		return VisibilityPrivate(), nil
	case visibilityPublic:
		return VisibilityPublic(), nil
	default:
		return Visibility{}, errors.New("invalid visibility: must be 'private' or 'public'")
	}
}

// Value returns the string representation.
func (v Visibility) Value() string { return v.value }

// IsPublic returns true if the visibility is public.
func (v Visibility) IsPublic() bool { return v.value == visibilityPublic }

// IsPrivate returns true if the visibility is private.
func (v Visibility) IsPrivate() bool { return v.value == visibilityPrivate }
