package gateway_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
