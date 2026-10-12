package gateway_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

type stubTokenSource struct {
	token *oauth2.Token
	err   error
}

func (s *stubTokenSource) Token() (*oauth2.Token, error) {
	return s.token, s.err
}

type recordingTransport struct {
	req  *http.Request
	resp *http.Response
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return t.resp, nil
}

func Test_serverlessAuthTransport_RoundTrip_shouldSetXServerlessAuthorizationHeader(t *testing.T) {
	t.Parallel()

	// given
	recorder := &recordingTransport{
		resp: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
	}
	tokenSrc := &stubTokenSource{
		token: &oauth2.Token{ //nolint:gosec
			AccessToken: "test-id-token-jwt",
			Expiry:      time.Now().Add(time.Hour),
		},
	}
	transport := gateway.NewServerlessAuthTransport(recorder, tokenSrc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer user-access-token")

	// when
	resp, err := transport.RoundTrip(req)

	// then
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer test-id-token-jwt", recorder.req.Header.Get("X-Serverless-Authorization"))
	assert.Equal(t, "Bearer user-access-token", recorder.req.Header.Get("Authorization"))
}

func Test_serverlessAuthTransport_RoundTrip_shouldReturnError_whenTokenSourceFails(t *testing.T) {
	t.Parallel()

	// given
	recorder := &recordingTransport{
		resp: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
	}
	tokenSrc := &stubTokenSource{
		err: errors.New("token error"),
	}
	transport := gateway.NewServerlessAuthTransport(recorder, tokenSrc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)

	// when
	resp, err := transport.RoundTrip(req) //nolint:bodyclose

	// then
	require.ErrorContains(t, err, "obtain ID token")
	assert.Nil(t, resp)
}

func Test_serverlessAuthTransport_RoundTrip_shouldReturnError_whenTokenIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	recorder := &recordingTransport{
		resp: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
	}
	tokenSrc := &stubTokenSource{
		token: &oauth2.Token{
			AccessToken: "",
		},
	}
	transport := gateway.NewServerlessAuthTransport(recorder, tokenSrc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/api", nil)
	require.NoError(t, err)

	// when
	resp, err := transport.RoundTrip(req) //nolint:bodyclose

	// then
	require.ErrorContains(t, err, "token is empty")
	assert.Nil(t, resp)
}

const newHTTPClientTimeout = 7 * time.Second

type appEnvTest struct {
	appEnv string
}

func appEnvsWithoutIDToken() []appEnvTest {
	return []appEnvTest{
		{appEnv: "local"},
		{appEnv: "test"},
	}
}

func Test_NewHTTPClient_shouldUseGivenTimeout_whenAppEnvIsLocalOrTest(t *testing.T) {
	t.Parallel()

	tests := appEnvsWithoutIDToken()
	for _, tt := range tests {
		t.Run(tt.appEnv, func(t *testing.T) {
			t.Parallel()

			// given, when
			client, err := gateway.NewHTTPClient(t.Context(), tt.appEnv, "https://example.com", newHTTPClientTimeout)

			// then
			require.NoError(t, err)
			assert.Equal(t, newHTTPClientTimeout, client.Timeout)
		})
	}
}

func Test_NewHTTPClient_shouldUseOtelTransport_whenAppEnvIsLocalOrTest(t *testing.T) {
	t.Parallel()

	tests := appEnvsWithoutIDToken()
	for _, tt := range tests {
		t.Run(tt.appEnv, func(t *testing.T) {
			t.Parallel()

			// given, when
			client, err := gateway.NewHTTPClient(t.Context(), tt.appEnv, "https://example.com", newHTTPClientTimeout)

			// then
			require.NoError(t, err)
			assert.IsType(t, &otelhttp.Transport{}, client.Transport)
		})
	}
}

func Test_NewHTTPClient_shouldNotSendServerlessAuthorization_whenAppEnvIsLocalOrTest(t *testing.T) {
	t.Parallel()

	tests := appEnvsWithoutIDToken()
	for _, tt := range tests {
		t.Run(tt.appEnv, func(t *testing.T) {
			t.Parallel()

			// given
			received := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received <- r.Header.Get("X-Serverless-Authorization")
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client, err := gateway.NewHTTPClient(t.Context(), tt.appEnv, server.URL, newHTTPClientTimeout)
			require.NoError(t, err)
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
			require.NoError(t, err)

			// when
			resp, err := client.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			// then
			assert.Empty(t, <-received)
		})
	}
}
