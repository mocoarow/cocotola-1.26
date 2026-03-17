// Package middleware provides Gin middleware for the cocotola-auth service.
// It includes an authentication middleware that validates JWT tokens from the
// Authorization header or cookies, sets user identity in the Gin context,
// and performs sliding token refresh for cookie-based sessions.
package middleware
