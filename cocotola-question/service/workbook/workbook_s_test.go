//go:build small

package workbook_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

type workbookFields struct {
	operatorID     string
	organizationID string
	targetID       string
	title          string
	description    string
	visibility     string
	language       string
}

func validWorkbookFields() workbookFields {
	return workbookFields{
		operatorID:     "op1",
		organizationID: "org1",
		targetID:       "target1",
		title:          "My Workbook",
		description:    "A description",
		visibility:     "private",
		language:       "en",
	}
}

func validWorkbookFieldVariants() []struct {
	name   string
	fields workbookFields
} {
	titleAtMax := validWorkbookFields()
	titleAtMax.title = strings.Repeat("a", 200)
	descriptionAtMax := validWorkbookFields()
	descriptionAtMax.description = strings.Repeat("d", 1000)
	emptyDescription := validWorkbookFields()
	emptyDescription.description = ""
	public := validWorkbookFields()
	public.visibility = "public"

	return []struct {
		name   string
		fields workbookFields
	}{
		{name: "all fields valid", fields: validWorkbookFields()},
		{name: "title at max length", fields: titleAtMax},
		{name: "description at max length", fields: descriptionAtMax},
		{name: "description empty", fields: emptyDescription},
		{name: "visibility public", fields: public},
	}
}

func invalidWorkbookFieldVariants() []struct {
	name   string
	fields workbookFields
} {
	with := func(set func(f *workbookFields)) workbookFields {
		f := validWorkbookFields()
		set(&f)
		return f
	}

	return []struct {
		name   string
		fields workbookFields
	}{
		{name: "operatorID empty", fields: with(func(f *workbookFields) { f.operatorID = "" })},
		{name: "organizationID empty", fields: with(func(f *workbookFields) { f.organizationID = "" })},
		{name: "targetID empty", fields: with(func(f *workbookFields) { f.targetID = "" })},
		{name: "title empty", fields: with(func(f *workbookFields) { f.title = "" })},
		{name: "title over max length", fields: with(func(f *workbookFields) { f.title = strings.Repeat("a", 201) })},
		{name: "description over max length", fields: with(func(f *workbookFields) { f.description = strings.Repeat("d", 1001) })},
		{name: "visibility empty", fields: with(func(f *workbookFields) { f.visibility = "" })},
		{name: "visibility unknown", fields: with(func(f *workbookFields) { f.visibility = "restricted" })},
		{name: "language empty", fields: with(func(f *workbookFields) { f.language = "" })},
		{name: "language one letter", fields: with(func(f *workbookFields) { f.language = "e" })},
		{name: "language three letters", fields: with(func(f *workbookFields) { f.language = "eng" })},
	}
}

func Test_NewCreateWorkbookInput_shouldReturnInput_whenFieldsAreValid(t *testing.T) {
	t.Parallel()

	for _, tt := range validWorkbookFieldVariants() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			f := tt.fields

			// when
			input, err := workbookservice.NewCreateWorkbookInput(f.operatorID, f.organizationID, f.targetID, f.title, f.description, f.visibility, f.language)

			// then
			require.NoError(t, err)
			assert.Equal(t, workbookservice.CreateWorkbookInput{
				OperatorID:     f.operatorID,
				OrganizationID: f.organizationID,
				SpaceID:        f.targetID,
				Title:          f.title,
				Description:    f.description,
				Visibility:     f.visibility,
				Language:       f.language,
			}, *input)
		})
	}
}

func Test_NewCreateWorkbookInput_shouldReturnErrInvalidArgument_whenFieldIsInvalid(t *testing.T) {
	t.Parallel()

	for _, tt := range invalidWorkbookFieldVariants() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			f := tt.fields

			// when
			_, err := workbookservice.NewCreateWorkbookInput(f.operatorID, f.organizationID, f.targetID, f.title, f.description, f.visibility, f.language)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUpdateWorkbookInput_shouldReturnInput_whenFieldsAreValid(t *testing.T) {
	t.Parallel()

	for _, tt := range validWorkbookFieldVariants() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			f := tt.fields

			// when
			input, err := workbookservice.NewUpdateWorkbookInput(f.operatorID, f.organizationID, f.targetID, f.title, f.description, f.visibility, f.language)

			// then
			require.NoError(t, err)
			assert.Equal(t, workbookservice.UpdateWorkbookInput{
				OperatorID:     f.operatorID,
				OrganizationID: f.organizationID,
				WorkbookID:     f.targetID,
				Title:          f.title,
				Description:    f.description,
				Visibility:     f.visibility,
				Language:       f.language,
			}, *input)
		})
	}
}

func Test_NewUpdateWorkbookInput_shouldReturnErrInvalidArgument_whenFieldIsInvalid(t *testing.T) {
	t.Parallel()

	for _, tt := range invalidWorkbookFieldVariants() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			f := tt.fields

			// when
			_, err := workbookservice.NewUpdateWorkbookInput(f.operatorID, f.organizationID, f.targetID, f.title, f.description, f.visibility, f.language)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewGetWorkbookInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := workbookservice.NewGetWorkbookInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	assert.Equal(t, workbookservice.GetWorkbookInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1"}, *input)
}

func Test_NewGetWorkbookInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
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
			_, err := workbookservice.NewGetWorkbookInput(tt.operatorID, tt.organizationID, tt.workbookID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewListWorkbooksInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := workbookservice.NewListWorkbooksInput("op1", "org1", "space1")

	// then
	require.NoError(t, err)
	assert.Equal(t, workbookservice.ListWorkbooksInput{OperatorID: "op1", OrganizationID: "org1", SpaceID: "space1"}, *input)
}

func Test_NewListWorkbooksInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		spaceID        string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", spaceID: "space1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", spaceID: "space1"},
		{name: "spaceID empty", operatorID: "op1", organizationID: "org1", spaceID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := workbookservice.NewListWorkbooksInput(tt.operatorID, tt.organizationID, tt.spaceID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewDeleteWorkbookInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := workbookservice.NewDeleteWorkbookInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	assert.Equal(t, workbookservice.DeleteWorkbookInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1"}, *input)
}

func Test_NewDeleteWorkbookInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
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
			_, err := workbookservice.NewDeleteWorkbookInput(tt.operatorID, tt.organizationID, tt.workbookID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
