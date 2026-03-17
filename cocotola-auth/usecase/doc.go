// Package usecase implements the application-level authentication logic for cocotola-auth.
// It follows the CQRS pattern, separating state-modifying Commands from read-only Queries.
// Each Command and Query is a focused struct with only the dependencies it needs.
package usecase
