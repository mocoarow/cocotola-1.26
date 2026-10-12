package group_test

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
	grouphandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/group"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var fixtureOperatorID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")

func setupCreateGroupRouter(t *testing.T, usecase *MockCreateGroupUsecase) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, fixtureOperatorID.String())
		c.Set(controller.ContextFieldOrganizationName{}, "cocotola")
		c.Next()
	})
	r.POST("/group", grouphandler.NewCreateGroupHandler(usecase).CreateGroup)
	return r
}

func serveCreateGroup(t *testing.T, usecaseErr error) *httptest.ResponseRecorder {
	t.Helper()
	usecase := NewMockCreateGroupUsecase(t)
	usecase.On("CreateGroup", mock.Anything, mock.Anything).Return(nil, usecaseErr)
	r := setupCreateGroupRouter(t, usecase)
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"name":"group-a"}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/group", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Test_CreateGroupHandler_shouldReturn409_whenConflictErrorIsReturned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		usecaseErr error
	}{
		{name: "duplicate group name", usecaseErr: fmt.Errorf("save group: %w", domain.ErrDuplicateEntry)},
		{name: "active group limit", usecaseErr: fmt.Errorf("add active group: %w", domain.ErrActiveGroupLimitReached)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			w := serveCreateGroup(t, tt.usecaseErr)

			// then
			assert.Equal(t, http.StatusConflict, w.Code)
		})
	}
}

func Test_CreateGroupHandler_shouldReturnFixedMessage_whenConflictErrorIsReturned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		usecaseErr error
		wantBody   string
	}{
		{
			name:       "duplicate group name",
			usecaseErr: fmt.Errorf("save group: %w", domain.ErrDuplicateEntry),
			wantBody:   `{"code":"conflict","message":"group already exists"}`,
		},
		{
			name:       "active group limit",
			usecaseErr: fmt.Errorf("add active group: %w", domain.ErrActiveGroupLimitReached),
			wantBody:   `{"code":"conflict","message":"active group limit reached"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			w := serveCreateGroup(t, tt.usecaseErr)

			// then
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
