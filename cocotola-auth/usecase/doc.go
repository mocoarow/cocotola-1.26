// Package usecase implements the application-level authentication logic for cocotola-auth.
// It orchestrates credential validation (AuthenticateCommand), JWT token parsing
// (AuthGetUserInfoQuery), and token refresh (AuthRefreshTokenQuery) through
// interface-based dependencies on token management capabilities.
package usecase
