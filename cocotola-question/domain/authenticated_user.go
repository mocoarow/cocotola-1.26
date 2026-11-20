package domain

// AuthenticatedUser is the caller identity resolved by cocotola-auth.
type AuthenticatedUser struct {
	UserID           string
	OrganizationName string
	Language         string
}
