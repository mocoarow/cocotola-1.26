// Package user implements user management use cases for cocotola-auth.
//
// Each use case is implemented as a separate Command struct
// following the CQRS pattern:
//
// Commands (state-modifying):
//   - CreateAppUserCommand: creates a new app user within an organization
//
// Command composes all commands via embedding.
package user
