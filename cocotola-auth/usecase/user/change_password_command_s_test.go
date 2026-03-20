package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
	userusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/user"
)

func Test_ChangePasswordCommand_ChangePassword_shouldSucceed_whenInputIsValid(t *testing.T) {
	t.Parallel()

	// given
	userID := 1
	newPassword := "newpassword123"
	hashedPassword := "$2a$10$newhash"

	user := domain.ReconstructAppUser(userID, 1, "user@example.com", "$2a$10$oldhash", true)

	finderMock := newMockappUserFinder(t)
	finderMock.On("FindByID", mock.Anything, userID).Return(user, nil)

	saverMock := newMockappUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", newPassword).Return(hashedPassword, nil)

	cmd := userusecase.NewChangePasswordCommand(finderMock, saverMock, hasherMock)
	input := &userservice.ChangePasswordInput{AppUserID: userID, NewPassword: newPassword}

	// when
	output, err := cmd.ChangePassword(context.Background(), input)

	// then
	require.NoError(t, err)
	require.Equal(t, userID, output.AppUserID)
	saverMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func Test_ChangePasswordCommand_ChangePassword_shouldReturnError_whenUserNotFound(t *testing.T) {
	t.Parallel()

	// given
	userID := 1
	newPassword := "newpassword123"

	finderMock := newMockappUserFinder(t)
	finderMock.On("FindByID", mock.Anything, userID).Return(nil, domain.ErrAppUserNotFound)

	saverMock := newMockappUserSaver(t)

	hasherMock := newMockpasswordHasher(t)

	cmd := userusecase.NewChangePasswordCommand(finderMock, saverMock, hasherMock)
	input := &userservice.ChangePasswordInput{AppUserID: userID, NewPassword: newPassword}

	// when
	output, err := cmd.ChangePassword(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrAppUserNotFound)
	require.Nil(t, output)
	saverMock.AssertNotCalled(t, "Save")
}

func Test_ChangePasswordCommand_ChangePassword_shouldReturnError_whenPasswordTooShort(t *testing.T) {
	t.Parallel()

	// given
	userID := 1
	newPassword := "short"

	user := domain.ReconstructAppUser(userID, 1, "user@example.com", "$2a$10$oldhash", true)

	finderMock := newMockappUserFinder(t)
	finderMock.On("FindByID", mock.Anything, userID).Return(user, nil)

	saverMock := newMockappUserSaver(t)

	hasherMock := newMockpasswordHasher(t)

	cmd := userusecase.NewChangePasswordCommand(finderMock, saverMock, hasherMock)
	input := &userservice.ChangePasswordInput{AppUserID: userID, NewPassword: newPassword}

	// when
	output, err := cmd.ChangePassword(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrPasswordTooShort)
	require.Nil(t, output)
	saverMock.AssertNotCalled(t, "Save")
}

func Test_ChangePasswordCommand_ChangePassword_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()

	// given
	userID := 1
	newPassword := "newpassword123"
	hashedPassword := "$2a$10$newhash"
	saveErr := errors.New("db error")

	user := domain.ReconstructAppUser(userID, 1, "user@example.com", "$2a$10$oldhash", true)

	finderMock := newMockappUserFinder(t)
	finderMock.On("FindByID", mock.Anything, userID).Return(user, nil)

	saverMock := newMockappUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", newPassword).Return(hashedPassword, nil)

	cmd := userusecase.NewChangePasswordCommand(finderMock, saverMock, hasherMock)
	input := &userservice.ChangePasswordInput{AppUserID: userID, NewPassword: newPassword}

	// when
	output, err := cmd.ChangePassword(context.Background(), input)

	// then
	require.ErrorIs(t, err, saveErr)
	require.Nil(t, output)
}
