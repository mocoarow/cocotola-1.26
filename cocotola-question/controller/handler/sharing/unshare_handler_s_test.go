package sharing_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

func Test_UnshareHandler_Unshare_shouldReturn204_whenUsecaseSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newSharingUsecases(t)
	usecases.unshare.On("Unshare", mock.Anything, &referenceservice.UnshareInput{
		OperatorID:     fixtureUserID,
		OrganizationID: fixtureOrganizationID,
		ReferenceID:    "ref-7",
	}).Return(nil).Once()
	r := initSharingRouter(ctx, t, usecases, fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

	// when
	w := serve(ctx, t, r, http.MethodDelete, "/api/v1/workbook/shared/ref-7")

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UnshareHandler_Unshare_shouldReturn401_whenOperatorIsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authMiddleware gin.HandlerFunc
	}{
		{name: "missing user ID", authMiddleware: fakeAuthMiddleware("", fixtureOrganizationID)},
		{name: "missing organization ID", authMiddleware: fakeAuthMiddleware(fixtureUserID, "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			r := initSharingRouter(ctx, t, newSharingUsecases(t), tt.authMiddleware)

			// when
			w := serve(ctx, t, r, http.MethodDelete, "/api/v1/workbook/shared/ref-7")

			// then
			require.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, api.ErrorResponse{Code: "unauthorized", Message: "Unauthorized"}, readErrorResponse(t, w))
		})
	}
}

func Test_UnshareHandler_Unshare_shouldReturn500_whenUsecaseFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newSharingUsecases(t)
	usecases.unshare.On("Unshare", mock.Anything, mock.Anything).Return(errors.New("firestore unavailable")).Once()
	r := initSharingRouter(ctx, t, usecases, fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

	// when
	w := serve(ctx, t, r, http.MethodDelete, "/api/v1/workbook/shared/ref-7")

	// then
	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, api.ErrorResponse{Code: "internal_server_error", Message: "Internal Server Error"}, readErrorResponse(t, w))
}
