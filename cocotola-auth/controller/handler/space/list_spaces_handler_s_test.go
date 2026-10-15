package space_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"

	spacehandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/space"
)

func getListSpaces(t *testing.T, usecase *MockListSpacesUsecase, userID string, organizationName string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldOrganizationName{}, organizationName)
		c.Next()
	})
	r.GET("/space", spacehandler.NewListSpacesHandler(usecase).ListSpaces)
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/space", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	return w
}

func serveListSpaces(t *testing.T, output *spaceservice.ListSpacesOutput, usecaseErr error) *httptest.ResponseRecorder {
	t.Helper()
	usecase := NewMockListSpacesUsecase(t)
	usecase.On("ListSpaces", mock.Anything, mock.Anything).Return(output, usecaseErr)
	return getListSpaces(t, usecase, fixtureOwnerID.String(), fixtureOrganizationName)
}

func Test_ListSpacesHandler_shouldReturn401_whenUserIDIsMissing(t *testing.T) {
	t.Parallel()

	// given, when
	w := getListSpaces(t, NewMockListSpacesUsecase(t), "", fixtureOrganizationName)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ListSpacesHandler_shouldReturn401_whenOrganizationNameIsMissing(t *testing.T) {
	t.Parallel()

	// given, when
	w := getListSpaces(t, NewMockListSpacesUsecase(t), fixtureOwnerID.String(), "")

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ListSpacesHandler_shouldPassOperatorAndOrganizationToUsecase_whenAuthenticated(t *testing.T) {
	t.Parallel()

	// given
	usecase := NewMockListSpacesUsecase(t)
	usecase.On("ListSpaces", mock.Anything, mock.MatchedBy(func(input *spaceservice.ListSpacesInput) bool {
		return input.OperatorID.Equal(fixtureOwnerID) && input.OrganizationName == fixtureOrganizationName
	})).Return(&spaceservice.ListSpacesOutput{}, nil)

	// when
	w := getListSpaces(t, usecase, fixtureOwnerID.String(), fixtureOrganizationName)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_ListSpacesHandler_shouldReturnSpaces_whenUsecaseSucceeds(t *testing.T) {
	t.Parallel()

	// given
	output := &spaceservice.ListSpacesOutput{
		Spaces: []spaceservice.Item{
			{
				SpaceID:        fixtureSpaceID,
				OrganizationID: fixtureOrgID,
				OwnerID:        fixtureOwnerID,
				KeyName:        "public-key",
				Name:           "Public Space",
				SpaceType:      domainspace.TypePublic().Value(),
				Deleted:        true,
			},
		},
	}

	// when
	w := serveListSpaces(t, output, nil)

	// then
	assert.JSONEq(t, fmt.Sprintf(`{"spaces":[{
		"spaceId":%q,
		"organizationId":%q,
		"ownerId":%q,
		"keyName":"public-key",
		"name":"Public Space",
		"spaceType":"public",
		"deleted":true
	}]}`, fixtureSpaceID.String(), fixtureOrgID.String(), fixtureOwnerID.String()), w.Body.String())
}

func Test_ListSpacesHandler_shouldReturnEmptyArray_whenNoSpacesExist(t *testing.T) {
	t.Parallel()

	// given, when
	w := serveListSpaces(t, &spaceservice.ListSpacesOutput{}, nil)

	// then
	assert.JSONEq(t, `{"spaces":[]}`, w.Body.String())
}

func Test_ListSpacesHandler_shouldMapUsecaseErrorToStatus_whenUsecaseFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "forbidden", err: fmt.Errorf("list spaces: %w", domain.ErrForbidden), wantStatus: http.StatusForbidden},
		{name: "organization not found", err: fmt.Errorf("list spaces: %w", domain.ErrOrganizationNotFound), wantStatus: http.StatusBadRequest},
		{name: "unexpected error", err: errors.New("database error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			w := serveListSpaces(t, nil, tt.err)

			// then
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
