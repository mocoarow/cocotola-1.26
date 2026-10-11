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

func Test_CreateUserHandler_shouldReturn409_whenLoginIDAlreadyExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockCreateUserUsecase(t)
	usecase.On("CreateAppUser", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("save app user: %w", domain.ErrDuplicateEntry))
	r := setupCreateUserRouter(t, usecase)
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"loginId":"user@example.com","password":"password123"}`)

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/user", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
}
