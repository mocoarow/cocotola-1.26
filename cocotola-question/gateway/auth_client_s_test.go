package gateway_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

const testAPIKey = "test-api-key"

func Test_AuthServicePolicyAdder_shouldSendPolicyRequest_whenValid(t *testing.T) {
	t.Parallel()

	// given
	var receivedBody string
	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-Service-Api-Key")
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		receivedBody = string(buf)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	adder := gateway.NewAuthServicePolicyAdder(server.URL, testAPIKey, server.Client())
	action, _ := domain.NewAction("create_question")
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	err = adder.AddPolicyForUser(context.Background(), "org-1", "user-1", action, resource, domain.EffectAllow())

	// then
	require.NoError(t, err)
	assert.Equal(t, testAPIKey, receivedAPIKey)
	assert.Contains(t, receivedBody, `"action":"create_question"`)
	assert.Contains(t, receivedBody, `"resource":"workbook:wb-123"`)
	assert.Contains(t, receivedBody, `"effect":"allow"`)
}

func Test_AuthServicePolicyAdder_shouldReturnError_whenServerReturnsError(t *testing.T) {
	t.Parallel()

	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	adder := gateway.NewAuthServicePolicyAdder(server.URL, "key", server.Client())
	action, _ := domain.NewAction("create_question")
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	err = adder.AddPolicyForUser(context.Background(), "org-1", "user-1", action, resource, domain.EffectAllow())

	// then
	require.ErrorContains(t, err, "status 500")
}

func Test_AuthServicePolicyAdder_shouldPostJSONToPolicyEndpointWithAPIKey_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	server, requests := newAuthServer(t, http.StatusNoContent, "")
	adder := gateway.NewAuthServicePolicyAdder(server.URL, testAPIKey, server.Client())
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	err = adder.AddPolicyForUser(context.Background(), "org-1", "user-1", domain.ActionCreateQuestion(), resource, domain.EffectAllow())

	// then
	require.NoError(t, err)
	assert.Equal(t, requestTarget{
		Method:      http.MethodPost,
		Path:        "/api/v1/internal/auth/authz/policy",
		APIKey:      testAPIKey,
		ContentType: "application/json",
	}, receivedRequest(t, requests).Target)
}

type requestTarget struct {
	Method      string
	Path        string
	Query       string
	APIKey      string
	ContentType string
}

type capturedRequest struct {
	Target requestTarget
	Body   string
}

func newAuthServer(t *testing.T, status int, body string) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	requests := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		requests <- capturedRequest{
			Target: requestTarget{
				Method:      r.Method,
				Path:        r.URL.EscapedPath(),
				Query:       r.URL.RawQuery,
				APIKey:      r.Header.Get("X-Service-Api-Key"),
				ContentType: r.Header.Get("Content-Type"),
			},
			Body: string(reqBody),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, requests
}

func receivedRequest(t *testing.T, requests <-chan capturedRequest) capturedRequest {
	t.Helper()

	select {
	case req := <-requests:
		return req
	default:
		require.FailNow(t, "auth service received no request")
		return capturedRequest{}
	}
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldReturnSpaceType_whenSpaceTypeIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		spaceType string
	}{
		{name: "public", spaceType: "public"},
		{name: "private", spaceType: "private"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			server, _ := newAuthServer(t, http.StatusOK, `{"spaceType":"`+tt.spaceType+`"}`)
			fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

			// when
			spaceType, err := fetcher.FetchSpaceType(context.Background(), "space-1")

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.spaceType, spaceType)
		})
	}
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldRequestEscapedSpacePathWithAPIKey_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	server, requests := newAuthServer(t, http.StatusOK, `{"spaceType":"public"}`)
	fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchSpaceType(context.Background(), "space/1")

	// then
	require.NoError(t, err)
	assert.Equal(t, requestTarget{
		Method: http.MethodGet,
		Path:   "/api/v1/internal/auth/space/space%2F1",
		APIKey: testAPIKey,
	}, receivedRequest(t, requests).Target)
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldReturnErrSpaceNotFound_whenAuthReturns404(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusNotFound, `{"error":"not found"}`)
	fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchSpaceType(context.Background(), "space-1")

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldReturnErrInvalidSpaceType_whenSpaceTypeIsUnknown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "unknown value", body: `{"spaceType":"shared"}`},
		{name: "empty value", body: `{"spaceType":""}`},
		{name: "missing field", body: `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			server, _ := newAuthServer(t, http.StatusOK, tt.body)
			fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

			// when
			_, err := fetcher.FetchSpaceType(context.Background(), "space-1")

			// then
			require.ErrorIs(t, err, domain.ErrInvalidSpaceType)
		})
	}
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldReturnStatusError_whenAuthReturnsServerError(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusInternalServerError, `{"error":"internal"}`)
	fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchSpaceType(context.Background(), "space-1")

	// then
	require.ErrorContains(t, err, "status 500")
}

func Test_AuthServiceSpaceTypeFetcher_FetchSpaceType_shouldReturnErrInvalidArgument_whenSpaceIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusOK, `{"spaceType":"public"}`)
	fetcher := gateway.NewAuthServiceSpaceTypeFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchSpaceType(context.Background(), "")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_AuthServiceMaxWorkbooksFetcher_FetchMaxWorkbooks_shouldReturnMaxWorkbooks_whenValueIsPositive(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusOK, `{"maxWorkbooks":7}`)
	fetcher := gateway.NewAuthServiceMaxWorkbooksFetcher(server.URL, testAPIKey, server.Client())

	// when
	maxWorkbooks, err := fetcher.FetchMaxWorkbooks(context.Background(), "user-1")

	// then
	require.NoError(t, err)
	assert.Equal(t, 7, maxWorkbooks)
}

func Test_AuthServiceMaxWorkbooksFetcher_FetchMaxWorkbooks_shouldRequestUserSettingWithAPIKey_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	server, requests := newAuthServer(t, http.StatusOK, `{"maxWorkbooks":7}`)
	fetcher := gateway.NewAuthServiceMaxWorkbooksFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchMaxWorkbooks(context.Background(), "user 1")

	// then
	require.NoError(t, err)
	assert.Equal(t, requestTarget{
		Method: http.MethodGet,
		Path:   "/api/v1/internal/auth/user-setting",
		Query:  "user_id=user+1",
		APIKey: testAPIKey,
	}, receivedRequest(t, requests).Target)
}

func Test_AuthServiceMaxWorkbooksFetcher_FetchMaxWorkbooks_shouldReturnError_whenValueIsNotPositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "zero", body: `{"maxWorkbooks":0}`},
		{name: "negative", body: `{"maxWorkbooks":-1}`},
		{name: "missing field", body: `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			server, _ := newAuthServer(t, http.StatusOK, tt.body)
			fetcher := gateway.NewAuthServiceMaxWorkbooksFetcher(server.URL, testAPIKey, server.Client())

			// when
			_, err := fetcher.FetchMaxWorkbooks(context.Background(), "user-1")

			// then
			require.ErrorContains(t, err, "invalid max workbooks value")
		})
	}
}

func Test_AuthServiceMaxWorkbooksFetcher_FetchMaxWorkbooks_shouldReturnStatusError_whenAuthReturnsServerError(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusInternalServerError, `{"error":"internal"}`)
	fetcher := gateway.NewAuthServiceMaxWorkbooksFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchMaxWorkbooks(context.Background(), "user-1")

	// then
	require.ErrorContains(t, err, "status 500")
}

func Test_AuthServiceMaxWorkbooksFetcher_FetchMaxWorkbooks_shouldReturnErrInvalidArgument_whenUserIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusOK, `{"maxWorkbooks":7}`)
	fetcher := gateway.NewAuthServiceMaxWorkbooksFetcher(server.URL, testAPIKey, server.Client())

	// when
	_, err := fetcher.FetchMaxWorkbooks(context.Background(), "")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_AuthServiceAuthorizationChecker_IsAllowed_shouldReturnAllowedFlag_whenAuthResponds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		allowed bool
	}{
		{name: "allowed", body: `{"allowed":true}`, allowed: true},
		{name: "denied", body: `{"allowed":false}`, allowed: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			server, _ := newAuthServer(t, http.StatusOK, tt.body)
			checker := gateway.NewAuthServiceAuthorizationChecker(server.URL, testAPIKey, server.Client())
			resource, err := domain.ResourceWorkbook("wb-123")
			require.NoError(t, err)

			// when
			allowed, err := checker.IsAllowed(context.Background(), "org-1", "user-1", domain.ActionViewWorkbook(), resource)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.allowed, allowed)
		})
	}
}

func Test_AuthServiceAuthorizationChecker_IsAllowed_shouldPostJSONToCheckEndpointWithAPIKey_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	server, requests := newAuthServer(t, http.StatusOK, `{"allowed":true}`)
	checker := gateway.NewAuthServiceAuthorizationChecker(server.URL, testAPIKey, server.Client())
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	_, err = checker.IsAllowed(context.Background(), "org-1", "user-1", domain.ActionViewWorkbook(), resource)

	// then
	require.NoError(t, err)
	assert.Equal(t, requestTarget{
		Method:      http.MethodPost,
		Path:        "/api/v1/internal/auth/authz/check",
		APIKey:      testAPIKey,
		ContentType: "application/json",
	}, receivedRequest(t, requests).Target)
}

func Test_AuthServiceAuthorizationChecker_IsAllowed_shouldSendCheckBody_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	server, requests := newAuthServer(t, http.StatusOK, `{"allowed":true}`)
	checker := gateway.NewAuthServiceAuthorizationChecker(server.URL, testAPIKey, server.Client())
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	_, err = checker.IsAllowed(context.Background(), "org-1", "user-1", domain.ActionViewWorkbook(), resource)

	// then
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"org":"org-1","user":"user-1","action":"view_workbook","resource":"workbook:wb-123"}`,
		receivedRequest(t, requests).Body)
}

func Test_AuthServiceAuthorizationChecker_IsAllowed_shouldReturnStatusError_whenAuthReturnsServerError(t *testing.T) {
	t.Parallel()

	// given
	server, _ := newAuthServer(t, http.StatusInternalServerError, `{"error":"internal"}`)
	checker := gateway.NewAuthServiceAuthorizationChecker(server.URL, testAPIKey, server.Client())
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	_, err = checker.IsAllowed(context.Background(), "org-1", "user-1", domain.ActionViewWorkbook(), resource)

	// then
	require.ErrorContains(t, err, "status 500")
}

func Test_AuthServiceAuthorizationChecker_IsAllowed_shouldReturnReadBodyError_whenErrorBodyIsTruncated(t *testing.T) {
	t.Parallel()

	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error"`))
	}))
	t.Cleanup(server.Close)
	checker := gateway.NewAuthServiceAuthorizationChecker(server.URL, testAPIKey, server.Client())
	resource, err := domain.ResourceWorkbook("wb-123")
	require.NoError(t, err)

	// when
	_, err = checker.IsAllowed(context.Background(), "org-1", "user-1", domain.ActionViewWorkbook(), resource)

	// then
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}
