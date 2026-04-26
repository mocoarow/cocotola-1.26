// Package domain provides shared domain constants used across cocotola microservices.
package domain

const (
	// SystemAppUserIDString is the UUID of the bootstrap "__system_admin" user.
	// This value is shared across all microservices to identify internal callers.
	SystemAppUserIDString = "00000000-0000-7000-8000-000000000002"
)
