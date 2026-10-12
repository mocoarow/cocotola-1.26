package user_test

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
	userhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/user"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var fixtureOperatorID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")

func setupCreateUserRouter(t *testing.T, usecase *MockCreateUserUsecase) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, fixtureOperatorID.String())
		c.Set(controller.ContextFieldOrganizationName{}, "cocotola")
		c.Next()
	})
	r.POST("/user", userhandler.NewCreateUserHandler(usecase).CreateUser)
	return r
}

func serveCreateUser(t *testing.T, usecaseErr error) *httptest.ResponseRecorder {
	t.Helper()
	usecase := NewMockCreateUserUsecase(t)
	usecase.On("CreateAppUser", mock.Anything, mock.Anything).Return(nil, usecaseErr)
	r := setupCreateUserRouter(t, usecase)
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"loginId":"user@example.com","password":"password123"}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/user", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Test_CreateUserHandler_shouldReturn409_whenConflictErrorIsReturned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		usecaseErr error
	}{
		{name: "duplicate login id", usecaseErr: fmt.Errorf("save app user: %w", domain.ErrDuplicateEntry)},
		{name: "active user limit", usecaseErr: fmt.Errorf("add active user: %w", domain.ErrActiveUserLimitReached)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			w := serveCreateUser(t, tt.usecaseErr)

			// then
			assert.Equal(t, http.StatusConflict, w.Code)
		})
	}
}

func Test_CreateUserHandler_shouldReturnFixedMessage_whenConflictErrorIsReturned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		usecaseErr error
		wantBody   string
	}{
		{
			name:       "duplicate login id",
			usecaseErr: fmt.Errorf("save app user: %w", domain.ErrDuplicateEntry),
			wantBody:   `{"code":"conflict","message":"user already exists"}`,
		},
		{
			name:       "active user limit",
			usecaseErr: fmt.Errorf("add active user: %w", domain.ErrActiveUserLimitReached),
			wantBody:   `{"code":"conflict","message":"active user limit reached"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			w := serveCreateUser(t, tt.usecaseErr)

			// then
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
