//go:build small

package reference_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

func Test_NewShareWorkbookInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := referenceservice.NewShareWorkbookInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	assert.Equal(t, referenceservice.ShareWorkbookInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1"}, *input)
}

func Test_NewShareWorkbookInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		workbookID     string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", workbookID: "wb1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", workbookID: "wb1"},
		{name: "workbookID empty", operatorID: "op1", organizationID: "org1", workbookID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := referenceservice.NewShareWorkbookInput(tt.operatorID, tt.organizationID, tt.workbookID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewListSharedInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := referenceservice.NewListSharedInput("op1", "org1")

	// then
	require.NoError(t, err)
	assert.Equal(t, referenceservice.ListSharedInput{OperatorID: "op1", OrganizationID: "org1"}, *input)
}

func Test_NewListSharedInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := referenceservice.NewListSharedInput(tt.operatorID, tt.organizationID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUnshareInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := referenceservice.NewUnshareInput("op1", "org1", "ref1")

	// then
	require.NoError(t, err)
	assert.Equal(t, referenceservice.UnshareInput{OperatorID: "op1", OrganizationID: "org1", ReferenceID: "ref1"}, *input)
}

func Test_NewUnshareInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		referenceID    string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", referenceID: "ref1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", referenceID: "ref1"},
		{name: "referenceID empty", operatorID: "op1", organizationID: "org1", referenceID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := referenceservice.NewUnshareInput(tt.operatorID, tt.organizationID, tt.referenceID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewListPublicInput_shouldReturnInput_whenFieldsAreValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language string
	}{
		{name: "language en", language: "en"},
		{name: "language ja", language: "ja"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			input, err := referenceservice.NewListPublicInput("op1", "org1", tt.language)

			// then
			require.NoError(t, err)
			assert.Equal(t, referenceservice.ListPublicInput{OperatorID: "op1", OrganizationID: "org1", Language: tt.language}, *input)
		})
	}
}

func Test_NewListPublicInput_shouldReturnErrInvalidArgument_whenFieldIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		language       string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", language: "en"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", language: "en"},
		{name: "language empty", operatorID: "op1", organizationID: "org1", language: ""},
		{name: "language one letter", operatorID: "op1", organizationID: "org1", language: "e"},
		{name: "language three letters", operatorID: "op1", organizationID: "org1", language: "eng"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := referenceservice.NewListPublicInput(tt.operatorID, tt.organizationID, tt.language)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
