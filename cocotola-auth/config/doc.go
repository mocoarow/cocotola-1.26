// Package config loads and validates the cocotola-auth application configuration.
// It reads an embedded config.yml file, expands environment variables with default
// value support, and unmarshals the result into typed structs covering server,
// database, authentication, tracing, and logging settings.
package config
