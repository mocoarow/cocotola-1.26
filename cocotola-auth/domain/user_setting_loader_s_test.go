package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var fixtureLoaderAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000574")

func Test_LoadOrInitUserSetting_shouldReturnStoredSetting_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	stored, err := domain.NewUserSetting(fixtureLoaderAppUserID, 5, "ja", 25, "America/Los_Angeles")
	require.NoError(t, err)
	finder := NewMockUserSettingFinder(t)
	finder.On("FindByAppUserID", mock.Anything, fixtureLoaderAppUserID).Return(stored, nil)

	// when
	got, err := domain.LoadOrInitUserSetting(ctx, finder, fixtureLoaderAppUserID)

	// then
	require.NoError(t, err)
	assert.Same(t, stored, got)
}

func Test_LoadOrInitUserSetting_shouldReturnDefaults_whenNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := NewMockUserSettingFinder(t)
	finder.On("FindByAppUserID", mock.Anything, fixtureLoaderAppUserID).Return(nil, domain.ErrUserSettingNotFound)
	expected, err := domain.NewDefaultUserSetting(fixtureLoaderAppUserID)
	require.NoError(t, err)

	// when
	got, err := domain.LoadOrInitUserSetting(ctx, finder, fixtureLoaderAppUserID)

	// then
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func Test_LoadOrInitUserSetting_shouldReturnFindError_whenFindFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	errFind := errors.New("db connection lost")
	finder := NewMockUserSettingFinder(t)
	finder.On("FindByAppUserID", mock.Anything, fixtureLoaderAppUserID).Return(nil, errFind)

	// when
	_, err := domain.LoadOrInitUserSetting(ctx, finder, fixtureLoaderAppUserID)

	// then
	require.ErrorIs(t, err, errFind)
}
