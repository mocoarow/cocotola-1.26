// Package auth implements authentication use cases for cocotola-auth.
//
// Each use case is implemented as a separate Command or Query struct
// following the CQRS pattern:
//
// Commands (state-modifying):
//   - PasswordAuthenticateCommand: verifies user credentials
//   - CreateSessionTokenCommand: creates a session token for cookie auth
//   - CreateTokenPairCommand: creates an access/refresh token pair for API auth
//   - ExtendSessionTokenCommand: extends session expiry (sliding window)
//   - RevokeSessionTokenCommand: revokes a session token
//   - RefreshAccessTokenCommand: issues a new access token from a refresh token
//   - RevokeTokenCommand: revokes an access or refresh token
//
// Queries (read-only):
//   - ValidateSessionTokenQuery: validates a session token and returns user info
//   - ValidateAccessTokenQuery: validates a JWT access token and returns user info
//
// Usecase composes all commands and queries via embedding, so it satisfies
// both the handler and middleware interfaces.
package auth
