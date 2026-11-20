//go:build small

package middleware_test

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

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const testAuthorization = "Bearer token"

func serveWithAuthMiddleware(t *testing.T, fetcher middleware.AuthMeFetcher, authorization string) (*httptest.ResponseRecorder, domain.AuthenticatedUser) {
	t.Helper()

	var captured domain.AuthenticatedUser
	r := gin.New()
	r.Use(middleware.NewAuthMiddleware(fetcher))
	r.GET("/protected", func(c *gin.Context) {
		captured = domain.AuthenticatedUser{
			UserID:           c.GetString(controller.ContextFieldUserID{}),
			OrganizationName: c.GetString(controller.ContextFieldOrganizationName{}),
			Language:         c.GetString(controller.ContextFieldUserLanguage{}),
		}
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, captured
}

func Test_NewAuthMiddleware_shouldReturn401_whenAuthorizationHeaderIsMissing(t *testing.T) {
	t.Parallel()

	// given
	fetcher := NewMockAuthMeFetcher(t)

	// when
	w, _ := serveWithAuthMiddleware(t, fetcher, "")

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_NewAuthMiddleware_shouldSetCallerInContext_whenFetchSucceeds(t *testing.T) {
	t.Parallel()

	// given
	me := domain.AuthenticatedUser{UserID: "user-42", OrganizationName: "org-42", Language: "ja"}
	fetcher := NewMockAuthMeFetcher(t)
	fetcher.On("FetchMe", mock.Anything, testAuthorization).Return(me, nil)

	// when
	_, captured := serveWithAuthMiddleware(t, fetcher, testAuthorization)

	// then
	assert.Equal(t, me, captured)
}

func Test_NewAuthMiddleware_shouldMapFetchErrorToStatus_whenFetchFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fetchErr   error
		wantStatus int
	}{
		{name: "unauthenticated", fetchErr: fmt.Errorf("fetch auth me: %w", domain.ErrUnauthenticated), wantStatus: http.StatusUnauthorized},
		{name: "auth service unavailable", fetchErr: fmt.Errorf("fetch auth me: %w", domain.ErrAuthServiceUnavailable), wantStatus: http.StatusBadGateway},
		{name: "unexpected error", fetchErr: errors.New("decode response"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			fetcher := NewMockAuthMeFetcher(t)
			fetcher.On("FetchMe", mock.Anything, testAuthorization).Return(domain.AuthenticatedUser{}, tt.fetchErr)

			// when
			w, _ := serveWithAuthMiddleware(t, fetcher, testAuthorization)

			// then
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
