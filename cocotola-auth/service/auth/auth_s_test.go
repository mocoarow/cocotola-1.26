package auth_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

const (
	fixtureLoginID      = "fixture-login-id"
	fixtureOrgName      = "fixture-org-name"
	fixturePassword     = "fixture-password"
	fixtureOpaqueString = "fixture-opaque-string"
	fixtureAccessToken  = "fixture-access-token"
	fixtureRefreshToken = "fixture-refresh-token"
	fixtureJWT          = "fixture-jwt"
)

var (
	fixtureUserID    = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000514")
	fixtureExpiresAt = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
)

type outcome struct {
	value any
	err   error
}

func result[T any](v *T, err error) outcome {
	return outcome{value: v, err: err}
}

func Test_constructors_shouldReturnValueWithGivenArguments_whenArgumentsAreValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      outcome
		expected any
	}{
		{"UserInfo", result(auth.NewUserInfo(fixtureUserID, fixtureLoginID, fixtureOrgName, fixtureExpiresAt)),
			&auth.UserInfo{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName, ExpiresAt: fixtureExpiresAt}},
		{"PasswordAuthenticateInput", result(auth.NewPasswordAuthenticateInput(fixtureLoginID, fixturePassword, fixtureOrgName)),
			&auth.PasswordAuthenticateInput{LoginID: fixtureLoginID, Password: fixturePassword, OrganizationName: fixtureOrgName}},
		{"PasswordAuthenticateOutput", result(auth.NewPasswordAuthenticateOutput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.PasswordAuthenticateOutput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"CreateSessionTokenInput", result(auth.NewCreateSessionTokenInput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.CreateSessionTokenInput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"CreateSessionTokenOutput", result(auth.NewCreateSessionTokenOutput(fixtureOpaqueString)),
			&auth.CreateSessionTokenOutput{RawToken: fixtureOpaqueString}},
		{"CreateTokenPairInput", result(auth.NewCreateTokenPairInput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.CreateTokenPairInput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"CreateTokenPairOutput", result(auth.NewCreateTokenPairOutput(fixtureAccessToken, fixtureRefreshToken)),
			&auth.CreateTokenPairOutput{AccessToken: fixtureAccessToken, RefreshToken: fixtureRefreshToken}},
		{"ValidateSessionTokenInput", result(auth.NewValidateSessionTokenInput(fixtureOpaqueString)),
			&auth.ValidateSessionTokenInput{RawToken: fixtureOpaqueString}},
		{"ValidateSessionTokenOutput", result(auth.NewValidateSessionTokenOutput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.ValidateSessionTokenOutput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"ValidateAccessTokenInput", result(auth.NewValidateAccessTokenInput(fixtureJWT)),
			&auth.ValidateAccessTokenInput{JWTString: fixtureJWT}},
		{"ValidateAccessTokenOutput", result(auth.NewValidateAccessTokenOutput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.ValidateAccessTokenOutput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"ExtendSessionTokenInput", result(auth.NewExtendSessionTokenInput(fixtureOpaqueString)),
			&auth.ExtendSessionTokenInput{RawToken: fixtureOpaqueString}},
		{"RevokeSessionTokenInput", result(auth.NewRevokeSessionTokenInput(fixtureOpaqueString)),
			&auth.RevokeSessionTokenInput{RawToken: fixtureOpaqueString}},
		{"RefreshAccessTokenInput", result(auth.NewRefreshAccessTokenInput(fixtureRefreshToken)),
			&auth.RefreshAccessTokenInput{RawRefreshToken: fixtureRefreshToken}},
		{"RefreshAccessTokenOutput", result(auth.NewRefreshAccessTokenOutput(fixtureAccessToken)),
			&auth.RefreshAccessTokenOutput{AccessToken: fixtureAccessToken}},
		{"GuestAuthenticateInput", result(auth.NewGuestAuthenticateInput(fixtureOrgName)),
			&auth.GuestAuthenticateInput{OrganizationName: fixtureOrgName}},
		{"GuestAuthenticateOutput", result(auth.NewGuestAuthenticateOutput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.GuestAuthenticateOutput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"SupabaseExchangeInput", result(auth.NewSupabaseExchangeInput(fixtureJWT, fixtureOrgName)),
			&auth.SupabaseExchangeInput{SupabaseJWT: fixtureJWT, OrganizationName: fixtureOrgName}},
		{"SupabaseExchangeOutput", result(auth.NewSupabaseExchangeOutput(fixtureUserID, fixtureLoginID, fixtureOrgName)),
			&auth.SupabaseExchangeOutput{UserID: fixtureUserID, LoginID: fixtureLoginID, OrganizationName: fixtureOrgName}},
		{"RevokeTokenInput", result(auth.NewRevokeTokenInput(fixtureOpaqueString)),
			&auth.RevokeTokenInput{Token: fixtureOpaqueString}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when: tt.got is built from valid fixtures

			// then
			require.NoError(t, tt.got.err)
			assert.Equal(t, tt.expected, tt.got.value)
		})
	}
}

func Test_constructors_shouldReturnErrInvalidArgument_whenUserIDIsZero(t *testing.T) {
	t.Parallel()

	var zero domain.AppUserID

	tests := []struct {
		name string
		got  outcome
	}{
		{"UserInfo", result(auth.NewUserInfo(zero, fixtureLoginID, fixtureOrgName, fixtureExpiresAt))},
		{"PasswordAuthenticateOutput", result(auth.NewPasswordAuthenticateOutput(zero, fixtureLoginID, fixtureOrgName))},
		{"CreateSessionTokenInput", result(auth.NewCreateSessionTokenInput(zero, fixtureLoginID, fixtureOrgName))},
		{"CreateTokenPairInput", result(auth.NewCreateTokenPairInput(zero, fixtureLoginID, fixtureOrgName))},
		{"ValidateSessionTokenOutput", result(auth.NewValidateSessionTokenOutput(zero, fixtureLoginID, fixtureOrgName))},
		{"ValidateAccessTokenOutput", result(auth.NewValidateAccessTokenOutput(zero, fixtureLoginID, fixtureOrgName))},
		{"GuestAuthenticateOutput", result(auth.NewGuestAuthenticateOutput(zero, fixtureLoginID, fixtureOrgName))},
		{"SupabaseExchangeOutput", result(auth.NewSupabaseExchangeOutput(zero, fixtureLoginID, fixtureOrgName))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when: tt.got is built with a zero user ID

			// then
			require.ErrorIs(t, tt.got.err, domain.ErrInvalidArgument)
		})
	}
}

func Test_constructors_shouldReturnValidationErrors_whenRequiredFieldIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  outcome
	}{
		{"UserInfo.LoginID", result(auth.NewUserInfo(fixtureUserID, "", fixtureOrgName, fixtureExpiresAt))},
		{"UserInfo.OrganizationName", result(auth.NewUserInfo(fixtureUserID, fixtureLoginID, "", fixtureExpiresAt))},
		{"UserInfo.ExpiresAt", result(auth.NewUserInfo(fixtureUserID, fixtureLoginID, fixtureOrgName, time.Time{}))},
		{"PasswordAuthenticateInput.LoginID", result(auth.NewPasswordAuthenticateInput("", fixturePassword, fixtureOrgName))},
		{"PasswordAuthenticateInput.Password", result(auth.NewPasswordAuthenticateInput(fixtureLoginID, "", fixtureOrgName))},
		{"PasswordAuthenticateInput.OrganizationName", result(auth.NewPasswordAuthenticateInput(fixtureLoginID, fixturePassword, ""))},
		{"PasswordAuthenticateOutput.LoginID", result(auth.NewPasswordAuthenticateOutput(fixtureUserID, "", fixtureOrgName))},
		{"PasswordAuthenticateOutput.OrganizationName", result(auth.NewPasswordAuthenticateOutput(fixtureUserID, fixtureLoginID, ""))},
		{"CreateSessionTokenInput.LoginID", result(auth.NewCreateSessionTokenInput(fixtureUserID, "", fixtureOrgName))},
		{"CreateSessionTokenInput.OrganizationName", result(auth.NewCreateSessionTokenInput(fixtureUserID, fixtureLoginID, ""))},
		{"CreateSessionTokenOutput.RawToken", result(auth.NewCreateSessionTokenOutput(""))},
		{"CreateTokenPairInput.LoginID", result(auth.NewCreateTokenPairInput(fixtureUserID, "", fixtureOrgName))},
		{"CreateTokenPairInput.OrganizationName", result(auth.NewCreateTokenPairInput(fixtureUserID, fixtureLoginID, ""))},
		{"CreateTokenPairOutput.AccessToken", result(auth.NewCreateTokenPairOutput("", fixtureRefreshToken))},
		{"CreateTokenPairOutput.RefreshToken", result(auth.NewCreateTokenPairOutput(fixtureAccessToken, ""))},
		{"ValidateSessionTokenInput.RawToken", result(auth.NewValidateSessionTokenInput(""))},
		{"ValidateSessionTokenOutput.LoginID", result(auth.NewValidateSessionTokenOutput(fixtureUserID, "", fixtureOrgName))},
		{"ValidateSessionTokenOutput.OrganizationName", result(auth.NewValidateSessionTokenOutput(fixtureUserID, fixtureLoginID, ""))},
		{"ValidateAccessTokenInput.JWTString", result(auth.NewValidateAccessTokenInput(""))},
		{"ValidateAccessTokenOutput.LoginID", result(auth.NewValidateAccessTokenOutput(fixtureUserID, "", fixtureOrgName))},
		{"ValidateAccessTokenOutput.OrganizationName", result(auth.NewValidateAccessTokenOutput(fixtureUserID, fixtureLoginID, ""))},
		{"ExtendSessionTokenInput.RawToken", result(auth.NewExtendSessionTokenInput(""))},
		{"RevokeSessionTokenInput.RawToken", result(auth.NewRevokeSessionTokenInput(""))},
		{"RefreshAccessTokenInput.RawRefreshToken", result(auth.NewRefreshAccessTokenInput(""))},
		{"RefreshAccessTokenOutput.AccessToken", result(auth.NewRefreshAccessTokenOutput(""))},
		{"GuestAuthenticateInput.OrganizationName", result(auth.NewGuestAuthenticateInput(""))},
		{"GuestAuthenticateOutput.LoginID", result(auth.NewGuestAuthenticateOutput(fixtureUserID, "", fixtureOrgName))},
		{"GuestAuthenticateOutput.OrganizationName", result(auth.NewGuestAuthenticateOutput(fixtureUserID, fixtureLoginID, ""))},
		{"SupabaseExchangeInput.SupabaseJWT", result(auth.NewSupabaseExchangeInput("", fixtureOrgName))},
		{"SupabaseExchangeInput.OrganizationName", result(auth.NewSupabaseExchangeInput(fixtureJWT, ""))},
		{"SupabaseExchangeOutput.LoginID", result(auth.NewSupabaseExchangeOutput(fixtureUserID, "", fixtureOrgName))},
		{"SupabaseExchangeOutput.OrganizationName", result(auth.NewSupabaseExchangeOutput(fixtureUserID, fixtureLoginID, ""))},
		{"RevokeTokenInput.Token", result(auth.NewRevokeTokenInput(""))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when: tt.got is built with one required field left empty

			// then
			var verrs validator.ValidationErrors
			require.ErrorAs(t, tt.got.err, &verrs)
		})
	}
}
