package user

// NewGuestLoginID returns the login ID for a guest user in the given organization.
func NewGuestLoginID(organizationName string) string {
	return "guest@@" + organizationName
}

// NewGuestUserName returns the display name for a guest user in the given organization.
func NewGuestUserName(organizationName string) string {
	return "Guest(" + organizationName + ")"
}
