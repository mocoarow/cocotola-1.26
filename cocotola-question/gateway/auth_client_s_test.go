package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
)

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

	adder := gateway.NewAuthServicePolicyAdder(server.URL, "test-api-key", server.Client())
	action, _ := domain.NewAction("create_question")
	resource := domain.ResourceWorkbook("wb-123")

	// when
	err := adder.AddPolicyForUser(context.Background(), "org-1", "user-1", action, resource, "allow")

	// then
	require.NoError(t, err)
	assert.Equal(t, "test-api-key", receivedAPIKey)
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
	resource := domain.ResourceWorkbook("wb-123")

	// when
	err := adder.AddPolicyForUser(context.Background(), "org-1", "user-1", action, resource, "allow")

	// then
	require.ErrorContains(t, err, "status 500")
}
