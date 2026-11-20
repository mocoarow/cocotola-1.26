package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const testAuthorization = "Bearer opaque-token"

type meRequest struct {
	Method        string
	Path          string
	Authorization string
	APIKey        string
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnAuthenticatedUser_whenAuthReturns200(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusOK)
	fetcher := gateway.NewAuthServiceMeFetcher(server.url, server.client, 0)

	// when
	me, err := fetcher.FetchMe(context.Background(), testAuthorization)

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.AuthenticatedUser{UserID: "user-42", OrganizationName: "org-42", Language: "ja"}, me)
}

func Test_AuthServiceMeFetcher_FetchMe_shouldSendExpectedMeRequest_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	requests := make(chan meRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- meRequest{
			Method:        r.Method,
			Path:          r.URL.EscapedPath(),
			Authorization: r.Header.Get("Authorization"),
			APIKey:        r.Header.Get("X-Service-Api-Key"),
		}
		_, _ = w.Write([]byte(`{"userId":"user-42"}`))
	}))
	t.Cleanup(server.Close)
	fetcher := gateway.NewAuthServiceMeFetcher(server.URL, server.Client(), 0)

	// when
	_, err := fetcher.FetchMe(context.Background(), testAuthorization)

	// then
	require.NoError(t, err)
	assert.Equal(t, meRequest{Method: http.MethodGet, Path: "/api/v1/auth/me", Authorization: testAuthorization, APIKey: ""}, <-requests)
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnUnauthenticated_whenAuthReturns401(t *testing.T) {
	t.Parallel()

	// given
	server := newMeAuthServer(t, http.StatusUnauthorized)
	fetcher := gateway.NewAuthServiceMeFetcher(server.url, server.client, 0)

	// when
	_, err := fetcher.FetchMe(context.Background(), testAuthorization)

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnAuthServiceUnavailable_whenAuthReturnsUnexpectedStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authStatus int
	}{
		{name: "internal server error", authStatus: http.StatusInternalServerError},
		{name: "service unavailable", authStatus: http.StatusServiceUnavailable},
		{name: "too many requests", authStatus: http.StatusTooManyRequests},
		{name: "forbidden", authStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			server := newMeAuthServer(t, tt.authStatus)
			fetcher := gateway.NewAuthServiceMeFetcher(server.url, server.client, 0)

			// when
			_, err := fetcher.FetchMe(context.Background(), testAuthorization)

			// then
			require.ErrorIs(t, err, domain.ErrAuthServiceUnavailable)
		})
	}
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnAuthServiceUnavailable_whenAuthIsUnreachable(t *testing.T) {
	t.Parallel()

	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !assert.True(t, ok) {
			return
		}
		conn, _, err := hijacker.Hijack()
		if !assert.NoError(t, err) {
			return
		}
		_ = conn.Close()
	}))
	t.Cleanup(server.Close)
	fetcher := gateway.NewAuthServiceMeFetcher(server.URL, server.Client(), 0)

	// when
	_, err := fetcher.FetchMe(context.Background(), testAuthorization)

	// then
	require.ErrorIs(t, err, domain.ErrAuthServiceUnavailable)
}

func Test_AuthServiceMeFetcher_FetchMe_shouldReturnDecodeError_whenResponseIsNotJSON(t *testing.T) {
	t.Parallel()

	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	t.Cleanup(server.Close)
	fetcher := gateway.NewAuthServiceMeFetcher(server.URL, server.Client(), 0)

	// when
	_, err := fetcher.FetchMe(context.Background(), testAuthorization)

	// then
	var syntaxErr *json.SyntaxError
	require.ErrorAs(t, err, &syntaxErr)
}
