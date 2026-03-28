package rbac_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	// when
	resource := rbac.ResourceUser(42)

	// then
	assert.Equal(t, "user:42", resource.Value())
}

func Test_ResourceGroup_shouldFormatWithGroupPrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := rbac.ResourceGroup(10)

	// then
	assert.Equal(t, "group:10", resource.Value())
}

func Test_ResourceSpace_shouldFormatWithSpacePrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := rbac.ResourceSpace(5)

	// then
	assert.Equal(t, "space:5", resource.Value())
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
		{name: "ActionViewUser", action: rbac.ActionViewUser(), want: "view_user"},
		{name: "ActionDisableUser", action: rbac.ActionDisableUser(), want: "disable_user"},
		{name: "ActionChangePassword", action: rbac.ActionChangePassword(), want: "change_password"},
		{name: "ActionCreateGroup", action: rbac.ActionCreateGroup(), want: "create_group"},
		{name: "ActionViewGroup", action: rbac.ActionViewGroup(), want: "view_group"},
		{name: "ActionDisableGroup", action: rbac.ActionDisableGroup(), want: "disable_group"},
		{name: "ActionAddUserToGroup", action: rbac.ActionAddUserToGroup(), want: "add_user_to_group"},
		{name: "ActionRemoveUserFromGroup", action: rbac.ActionRemoveUserFromGroup(), want: "remove_user_from_group"},
		{name: "ActionCreateSpace", action: rbac.ActionCreateSpace(), want: "create_space"},
		{name: "ActionViewSpace", action: rbac.ActionViewSpace(), want: "view_space"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Value())
		})
	}
}
