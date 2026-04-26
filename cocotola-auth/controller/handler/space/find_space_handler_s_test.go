package space_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"

	spacehandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/space"
)

var (
	fixtureSpaceID = domain.MustParseSpaceID("11111111-1111-7111-8111-111111111111")
	fixtureOrgID   = domain.MustParseOrganizationID("22222222-2222-7222-8222-222222222222")
	fixtureOwnerID = domain.MustParseAppUserID("33333333-3333-7333-8333-333333333333")
)

func setupFindSpaceRouter(t *testing.T, usecase *MockFindSpaceUsecase) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := spacehandler.NewFindSpaceHandler(usecase)
	r.GET("/internal/auth/space/:spaceId", handler.FindSpace)
	return r
}

func Test_FindSpaceHandler_shouldReturn200_whenSpaceExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	output := &spaceservice.FindSpaceOutput{
		Item: spaceservice.Item{
			SpaceID:        fixtureSpaceID,
			OrganizationID: fixtureOrgID,
			OwnerID:        fixtureOwnerID,
			KeyName:        "test-key",
			Name:           "Test Space",
			SpaceType:      domainspace.TypePrivate().Value(),
			Deleted:        false,
		},
	}
	usecase := NewMockFindSpaceUsecase(t)
	usecase.On("FindSpace", mock.Anything, mock.MatchedBy(func(input *spaceservice.FindSpaceInput) bool {
		return input.SpaceID.Equal(fixtureSpaceID)
	})).Return(output, nil)

	r := setupFindSpaceRouter(t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/auth/space/"+fixtureSpaceID.String(), nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_FindSpaceHandler_shouldReturn400_whenSpaceIDIsInvalidUUID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockFindSpaceUsecase(t)
	r := setupFindSpaceRouter(t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/auth/space/not-a-uuid", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_FindSpaceHandler_shouldReturn404_whenSpaceNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockFindSpaceUsecase(t)
	usecase.On("FindSpace", mock.Anything, mock.Anything).Return(nil, domain.ErrSpaceNotFound)

	r := setupFindSpaceRouter(t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/auth/space/"+fixtureSpaceID.String(), nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_FindSpaceHandler_shouldReturn500_whenUsecaseFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockFindSpaceUsecase(t)
	usecase.On("FindSpace", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	r := setupFindSpaceRouter(t, usecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/auth/space/"+fixtureSpaceID.String(), nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
