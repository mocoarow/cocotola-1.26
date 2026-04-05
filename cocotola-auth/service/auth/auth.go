// Package auth provides service-layer types for authentication input/output validation.
package auth

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// UserInfo represents an authenticated user's identity.
// Used internally by gateway-layer interfaces (JWTManager, UserAuthenticator).
type UserInfo struct {
	UserID           int       `validate:"required,gt=0"`
	LoginID          string    `validate:"required"`
	OrganizationName string    `validate:"required"`
	ExpiresAt        time.Time `validate:"required"`
}

// NewUserInfo creates a validated UserInfo.
func NewUserInfo(userID int, loginID string, organizationName string, expiresAt time.Time) (*UserInfo, error) {
	m := &UserInfo{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
		ExpiresAt:        expiresAt,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate user info: %w", err)
	}
	return m, nil
}

// --- PasswordAuthenticate ---

// PasswordAuthenticateInput holds the login credentials for password authentication.
type PasswordAuthenticateInput struct {
	LoginID          string `validate:"required"`
	Password         string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewPasswordAuthenticateInput creates a validated PasswordAuthenticateInput.
func NewPasswordAuthenticateInput(loginID string, password string, organizationName string) (*PasswordAuthenticateInput, error) {
	m := &PasswordAuthenticateInput{
		LoginID:          loginID,
		Password:         password,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate password authenticate input: %w", err)
	}
	return m, nil
}

// PasswordAuthenticateOutput holds the authenticated user's identity.
type PasswordAuthenticateOutput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewPasswordAuthenticateOutput creates a validated PasswordAuthenticateOutput.
func NewPasswordAuthenticateOutput(userID int, loginID string, organizationName string) (*PasswordAuthenticateOutput, error) {
	m := &PasswordAuthenticateOutput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate password authenticate output: %w", err)
	}
	return m, nil
}

// --- CreateSessionToken ---

// CreateSessionTokenInput holds the parameters for creating a session token (cookie auth).
type CreateSessionTokenInput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewCreateSessionTokenInput creates a validated CreateSessionTokenInput.
func NewCreateSessionTokenInput(userID int, loginID string, organizationName string) (*CreateSessionTokenInput, error) {
	m := &CreateSessionTokenInput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create session token input: %w", err)
	}
	return m, nil
}

// CreateSessionTokenOutput holds the raw session token string to be set in a cookie.
type CreateSessionTokenOutput struct {
	RawToken string `validate:"required"`
}

// NewCreateSessionTokenOutput creates a validated CreateSessionTokenOutput.
func NewCreateSessionTokenOutput(rawToken string) (*CreateSessionTokenOutput, error) {
	m := &CreateSessionTokenOutput{
		RawToken: rawToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create session token output: %w", err)
	}
	return m, nil
}

// --- CreateTokenPair ---

// CreateTokenPairInput holds the parameters for creating an access/refresh token pair (API auth).
type CreateTokenPairInput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewCreateTokenPairInput creates a validated CreateTokenPairInput.
func NewCreateTokenPairInput(userID int, loginID string, organizationName string) (*CreateTokenPairInput, error) {
	m := &CreateTokenPairInput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create token pair input: %w", err)
	}
	return m, nil
}

// CreateTokenPairOutput holds the JWT access token and opaque refresh token.
type CreateTokenPairOutput struct {
	AccessToken  string `validate:"required"`
	RefreshToken string `validate:"required"`
}

// NewCreateTokenPairOutput creates a validated CreateTokenPairOutput.
func NewCreateTokenPairOutput(accessToken string, refreshToken string) (*CreateTokenPairOutput, error) {
	m := &CreateTokenPairOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate create token pair output: %w", err)
	}
	return m, nil
}

// --- ValidateSessionToken ---

// ValidateSessionTokenInput holds the raw session token to validate.
type ValidateSessionTokenInput struct {
	RawToken string `validate:"required"`
}

// NewValidateSessionTokenInput creates a validated ValidateSessionTokenInput.
func NewValidateSessionTokenInput(rawToken string) (*ValidateSessionTokenInput, error) {
	m := &ValidateSessionTokenInput{
		RawToken: rawToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("new validate session token input: %w", err)
	}
	return m, nil
}

// ValidateSessionTokenOutput holds the validated session token's user info.
type ValidateSessionTokenOutput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewValidateSessionTokenOutput creates a validated ValidateSessionTokenOutput.
func NewValidateSessionTokenOutput(userID int, loginID string, organizationName string) (*ValidateSessionTokenOutput, error) {
	m := &ValidateSessionTokenOutput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("new validate session token output: %w", err)
	}
	return m, nil
}

// --- ValidateAccessToken ---

// ValidateAccessTokenInput holds the JWT string to validate.
type ValidateAccessTokenInput struct {
	JWTString string `validate:"required"`
}

// NewValidateAccessTokenInput creates a validated ValidateAccessTokenInput.
func NewValidateAccessTokenInput(jwtString string) (*ValidateAccessTokenInput, error) {
	m := &ValidateAccessTokenInput{
		JWTString: jwtString,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("new validate access token input: %w", err)
	}
	return m, nil
}

// ValidateAccessTokenOutput holds the validated access token's user info.
type ValidateAccessTokenOutput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewValidateAccessTokenOutput creates a validated ValidateAccessTokenOutput.
func NewValidateAccessTokenOutput(userID int, loginID string, organizationName string) (*ValidateAccessTokenOutput, error) {
	m := &ValidateAccessTokenOutput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("new validate access token output: %w", err)
	}
	return m, nil
}

// --- ExtendSessionToken ---

// ExtendSessionTokenInput holds the raw session token to extend.
type ExtendSessionTokenInput struct {
	RawToken string `validate:"required"`
}

// NewExtendSessionTokenInput creates a validated ExtendSessionTokenInput.
func NewExtendSessionTokenInput(rawToken string) (*ExtendSessionTokenInput, error) {
	m := &ExtendSessionTokenInput{
		RawToken: rawToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate extend session token input: %w", err)
	}
	return m, nil
}

// --- RevokeSessionToken ---

// RevokeSessionTokenInput holds the raw session token to revoke.
type RevokeSessionTokenInput struct {
	RawToken string `validate:"required"`
}

// NewRevokeSessionTokenInput creates a validated RevokeSessionTokenInput.
func NewRevokeSessionTokenInput(rawToken string) (*RevokeSessionTokenInput, error) {
	m := &RevokeSessionTokenInput{
		RawToken: rawToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate revoke session token input: %w", err)
	}
	return m, nil
}

// --- RefreshAccessToken ---

// RefreshAccessTokenInput holds the raw refresh token for token refresh.
type RefreshAccessTokenInput struct {
	RawRefreshToken string `validate:"required"`
}

// NewRefreshAccessTokenInput creates a validated RefreshAccessTokenInput.
func NewRefreshAccessTokenInput(rawRefreshToken string) (*RefreshAccessTokenInput, error) {
	m := &RefreshAccessTokenInput{
		RawRefreshToken: rawRefreshToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate refresh access token input: %w", err)
	}
	return m, nil
}

// RefreshAccessTokenOutput holds a newly issued JWT access token.
type RefreshAccessTokenOutput struct {
	AccessToken string `validate:"required"`
}

// NewRefreshAccessTokenOutput creates a validated RefreshAccessTokenOutput.
func NewRefreshAccessTokenOutput(accessToken string) (*RefreshAccessTokenOutput, error) {
	m := &RefreshAccessTokenOutput{
		AccessToken: accessToken,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate refresh access token output: %w", err)
	}
	return m, nil
}

// --- GuestAuthenticate ---

// GuestAuthenticateInput holds the organization name for guest authentication.
type GuestAuthenticateInput struct {
	OrganizationName string `validate:"required"`
}

// NewGuestAuthenticateInput creates a validated GuestAuthenticateInput.
func NewGuestAuthenticateInput(organizationName string) (*GuestAuthenticateInput, error) {
	m := &GuestAuthenticateInput{
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate guest authenticate input: %w", err)
	}
	return m, nil
}

// GuestAuthenticateOutput holds the authenticated guest user's identity.
type GuestAuthenticateOutput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewGuestAuthenticateOutput creates a validated GuestAuthenticateOutput.
func NewGuestAuthenticateOutput(userID int, loginID string, organizationName string) (*GuestAuthenticateOutput, error) {
	m := &GuestAuthenticateOutput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate guest authenticate output: %w", err)
	}
	return m, nil
}

// --- SupabaseExchange ---

// SupabaseExchangeInput holds the Supabase JWT and organization name for token exchange.
type SupabaseExchangeInput struct {
	SupabaseJWT      string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewSupabaseExchangeInput creates a validated SupabaseExchangeInput.
func NewSupabaseExchangeInput(supabaseJWT string, organizationName string) (*SupabaseExchangeInput, error) {
	m := &SupabaseExchangeInput{
		SupabaseJWT:      supabaseJWT,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate supabase exchange input: %w", err)
	}
	return m, nil
}

// SupabaseExchangeOutput holds the authenticated user's identity after Supabase token exchange.
type SupabaseExchangeOutput struct {
	UserID           int    `validate:"required,gt=0"`
	LoginID          string `validate:"required"`
	OrganizationName string `validate:"required"`
}

// NewSupabaseExchangeOutput creates a validated SupabaseExchangeOutput.
func NewSupabaseExchangeOutput(userID int, loginID string, organizationName string) (*SupabaseExchangeOutput, error) {
	m := &SupabaseExchangeOutput{
		UserID:           userID,
		LoginID:          loginID,
		OrganizationName: organizationName,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate supabase exchange output: %w", err)
	}
	return m, nil
}

// --- RevokeToken ---

// RevokeTokenInput holds the token string to revoke.
type RevokeTokenInput struct {
	Token string `validate:"required"`
}

// NewRevokeTokenInput creates a validated RevokeTokenInput.
func NewRevokeTokenInput(token string) (*RevokeTokenInput, error) {
	m := &RevokeTokenInput{
		Token: token,
	}
	if err := domain.ValidateStruct(m); err != nil {
		return nil, fmt.Errorf("validate revoke token input: %w", err)
	}
	return m, nil
}
