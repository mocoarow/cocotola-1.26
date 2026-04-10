package rbac_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
)

func Test_NewRBACAction_shouldReturnAction_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "create_user"

	// when
	action, err := rbac.NewAction(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, action.Value())
}

func Test_NewRBACAction_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := rbac.NewAction("")

	// then
	require.Error(t, err)
}

func Test_NewRBACResource_shouldReturnResource_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "user:1"

	// when
	resource, err := rbac.NewResource(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, resource.Value())
}

func Test_NewRBACResource_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := rbac.NewResource("")

	// then
	require.Error(t, err)
}

func Test_ResourceUser_shouldFormatWithUserPrefix(t *testing.T) {
	t.Parallel()

	// given
	userID := domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")

	// when
	resource := rbac.ResourceUser(userID)

	// then
	assert.Equal(t, "user:"+userID.String(), resource.Value())
}

func Test_ResourceGroup_shouldFormatWithGroupPrefix(t *testing.T) {
	t.Parallel()

	// given
	groupID := domain.MustParseGroupID("00000000-0000-7000-8000-000000000010")

	// when
	resource := rbac.ResourceGroup(groupID)

	// then
	assert.Equal(t, "group:"+groupID.String(), resource.Value())
}

func Test_ResourceSpace_shouldFormatWithSpacePrefix(t *testing.T) {
	t.Parallel()

	// given
	spaceID := domain.MustParseSpaceID("00000000-0000-7000-8000-000000000005")

	// when
	resource := rbac.ResourceSpace(spaceID)

	// then
	assert.Equal(t, "space:"+spaceID.String(), resource.Value())
}

func Test_ResourceWorkbook_shouldFormatWithWorkbookPrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := rbac.ResourceWorkbook("abc123")

	// then
	assert.Equal(t, "workbook:abc123", resource.Value())
}

func Test_ResourceQuestion_shouldFormatWithQuestionPrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := rbac.ResourceQuestion("q456")

	// then
	assert.Equal(t, "question:q456", resource.Value())
}

func Test_ResourceAny_shouldReturnWildcard(t *testing.T) {
	t.Parallel()

	// when
	resource := rbac.ResourceAny()

	// then
	assert.Equal(t, "*", resource.Value())
}

func Test_NewRBACGroup_shouldReturnGroup_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "admin"

	// when
	group, err := rbac.NewGroup(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, group.Value())
}

func Test_NewRBACGroup_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := rbac.NewGroup("")

	// then
	require.Error(t, err)
}

func Test_EffectAllow_shouldReturnAllow(t *testing.T) {
	t.Parallel()

	// then
	assert.Equal(t, "allow", rbac.EffectAllow().Value())
}

func Test_EffectDeny_shouldReturnDeny(t *testing.T) {
	t.Parallel()

	// then
	assert.Equal(t, "deny", rbac.EffectDeny().Value())
}

func Test_PredefinedActions_shouldHaveCorrectValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action rbac.Action
		want   string
	}{
		{name: "ActionCreateUser", action: rbac.ActionCreateUser(), want: "create_user"},
		{name: "ActionListUsers", action: rbac.ActionListUsers(), want: "list_users"},
		{name: "ActionViewUser", action: rbac.ActionViewUser(), want: "view_user"},
		{name: "ActionDisableUser", action: rbac.ActionDisableUser(), want: "disable_user"},
		{name: "ActionChangePassword", action: rbac.ActionChangePassword(), want: "change_password"},
		{name: "ActionCreateGroup", action: rbac.ActionCreateGroup(), want: "create_group"},
		{name: "ActionListGroups", action: rbac.ActionListGroups(), want: "list_groups"},
		{name: "ActionViewGroup", action: rbac.ActionViewGroup(), want: "view_group"},
		{name: "ActionDisableGroup", action: rbac.ActionDisableGroup(), want: "disable_group"},
		{name: "ActionAddUserToGroup", action: rbac.ActionAddUserToGroup(), want: "add_user_to_group"},
		{name: "ActionRemoveUserFromGroup", action: rbac.ActionRemoveUserFromGroup(), want: "remove_user_from_group"},
		{name: "ActionCreateSpace", action: rbac.ActionCreateSpace(), want: "create_space"},
		{name: "ActionListSpaces", action: rbac.ActionListSpaces(), want: "list_spaces"},
		{name: "ActionViewSpace", action: rbac.ActionViewSpace(), want: "view_space"},
		{name: "ActionCreateWorkbook", action: rbac.ActionCreateWorkbook(), want: "create_workbook"},
		{name: "ActionViewWorkbook", action: rbac.ActionViewWorkbook(), want: "view_workbook"},
		{name: "ActionUpdateWorkbook", action: rbac.ActionUpdateWorkbook(), want: "update_workbook"},
		{name: "ActionDeleteWorkbook", action: rbac.ActionDeleteWorkbook(), want: "delete_workbook"},
		{name: "ActionImportWorkbook", action: rbac.ActionImportWorkbook(), want: "import_workbook"},
		{name: "ActionCreateQuestion", action: rbac.ActionCreateQuestion(), want: "create_question"},
		{name: "ActionUpdateQuestion", action: rbac.ActionUpdateQuestion(), want: "update_question"},
		{name: "ActionDeleteQuestion", action: rbac.ActionDeleteQuestion(), want: "delete_question"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Value())
		})
	}
}
