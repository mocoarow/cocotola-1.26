// Package domain defines the core types and validation rules for the cocotola-auth service.
// It contains authentication input/output value objects, user identity types,
// token refresh types, and sentinel errors such as ErrUnauthenticated.
// Domain types enforce their invariants through constructor functions that validate
// struct fields using go-playground/validator tags.
package domain
