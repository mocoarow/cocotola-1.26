package space_test

import (
	"bytes"
	"context"
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

	spacehandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/space"
)

func postCreateSpace(t *testing.T, usecase *MockCreateSpaceUsecase, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, fixtureOwnerID.String())
		c.Set(controller.ContextFieldOrganizationName{}, fixtureOrganizationName)
		c.Next()
	})
	r.POST("/space", spacehandler.NewCreateSpaceHandler(usecase).CreateSpace)
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/space", bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func serveCreateSpace(t *testing.T, usecaseErr error) *httptest.ResponseRecorder {
	t.Helper()
	usecase := NewMockCreateSpaceUsecase(t)
	usecase.On("CreateSpace", mock.Anything, mock.Anything).Return(nil, usecaseErr)
	return postCreateSpace(t, usecase, `{"name":"Public","spaceType":"public"}`)
}

func Test_CreateSpaceHandler_shouldReturn400_whenSpaceTypeIsPrivate(t *testing.T) {
	t.Parallel()

	// given, when
	w := postCreateSpace(t, NewMockCreateSpaceUsecase(t), `{"name":"Private","spaceType":"private"}`)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_CreateSpaceHandler_shouldReturn409_whenSpaceAlreadyExists(t *testing.T) {
	t.Parallel()

	// given, when
	w := serveCreateSpace(t, fmt.Errorf("provision space: %w", domain.ErrDuplicateEntry))

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
}

func Test_CreateSpaceHandler_shouldReturnFixedMessage_whenSpaceAlreadyExists(t *testing.T) {
	t.Parallel()

	// given, when
	w := serveCreateSpace(t, fmt.Errorf("provision space: %w", domain.ErrDuplicateEntry))

	// then
	assert.JSONEq(t, `{"code":"conflict","message":"space already exists"}`, w.Body.String())
}
