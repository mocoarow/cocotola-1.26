package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewRBACAction_shouldReturnAction_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "create_user"

	// when
	action, err := domain.NewRBACAction(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, action.Value())
}

func Test_NewRBACAction_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewRBACAction("")

	// then
	require.Error(t, err)
}

func Test_NewRBACResource_shouldReturnResource_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "user:1"

	// when
	resource, err := domain.NewRBACResource(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, resource.Value())
}

func Test_NewRBACResource_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewRBACResource("")

	// then
	require.Error(t, err)
}

func Test_ResourceUser_shouldFormatWithUserPrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := domain.ResourceUser(42)

	// then
	assert.Equal(t, "user:42", resource.Value())
}

func Test_ResourceGroup_shouldFormatWithGroupPrefix(t *testing.T) {
	t.Parallel()

	// when
	resource := domain.ResourceGroup(10)

	// then
	assert.Equal(t, "group:10", resource.Value())
}

func Test_ResourceAny_shouldReturnWildcard(t *testing.T) {
	t.Parallel()

	// when
	resource := domain.ResourceAny()

	// then
	assert.Equal(t, "*", resource.Value())
}

func Test_NewRBACGroup_shouldReturnGroup_whenValueIsValid(t *testing.T) {
	t.Parallel()

	// given
	value := "admin"

	// when
	group, err := domain.NewRBACGroup(value)

	// then
	require.NoError(t, err)
	assert.Equal(t, value, group.Value())
}

func Test_NewRBACGroup_shouldReturnError_whenValueIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := domain.NewRBACGroup("")

	// then
	require.Error(t, err)
}

func Test_EffectAllow_shouldReturnAllow(t *testing.T) {
	t.Parallel()

	// then
	assert.Equal(t, "allow", domain.EffectAllow().Value())
}

func Test_EffectDeny_shouldReturnDeny(t *testing.T) {
	t.Parallel()

	// then
	assert.Equal(t, "deny", domain.EffectDeny().Value())
}

func Test_PredefinedActions_shouldHaveCorrectValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action domain.RBACAction
		want   string
	}{
		{name: "ActionCreateUser", action: domain.ActionCreateUser(), want: "create_user"},
		{name: "ActionViewUser", action: domain.ActionViewUser(), want: "view_user"},
		{name: "ActionDisableUser", action: domain.ActionDisableUser(), want: "disable_user"},
		{name: "ActionChangePassword", action: domain.ActionChangePassword(), want: "change_password"},
		{name: "ActionCreateGroup", action: domain.ActionCreateGroup(), want: "create_group"},
		{name: "ActionViewGroup", action: domain.ActionViewGroup(), want: "view_group"},
		{name: "ActionDisableGroup", action: domain.ActionDisableGroup(), want: "disable_group"},
		{name: "ActionAddUserToGroup", action: domain.ActionAddUserToGroup(), want: "add_user_to_group"},
		{name: "ActionRemoveUserFromGroup", action: domain.ActionRemoveUserFromGroup(), want: "remove_user_from_group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Value())
		})
	}
}
