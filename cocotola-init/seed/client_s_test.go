//go:build small

package seed_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

const (
	testAPIKey       = "test-api-key"
	testRequestOrgID = "org-1"
)

// recordedCall captures one inbound HTTP request the client makes during a test.
type recordedCall struct {
	Method  string
	Path    string
	Query   string
	Headers http.Header
	Body    []byte
}

func newRecorder() (*httptest.Server, *[]recordedCall) {
	var (
		mu    sync.Mutex
		calls []recordedCall
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/internal/workbook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, recordedCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			Headers: r.Header.Clone(),
			Body:    body,
		})
		mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"workbooks":[{"workbookId":"wb-1","title":"Existing","description":"desc [seedKey:k1]"}]}`))
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"workbookId":"wb-new"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/internal/workbook/wb-1/question", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, recordedCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    body,
		})
		mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"questions":[]}`))
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"questionId":"q-new"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(mux)
	return srv, &calls
}

func newClient(srv *httptest.Server) *seed.QuestionAPIClient {
	return seed.NewQuestionAPIClient(srv.URL, testAPIKey, srv.Client())
}

func Test_QuestionAPIClient_ListWorkbooks_shouldSendApiKeyAndOrgIDHeaders(t *testing.T) {
	t.Parallel()
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	out, err := client.ListWorkbooks(context.Background(), testRequestOrgID, "space-1")

	// then
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	c := (*calls)[0]
	assert.Equal(t, http.MethodGet, c.Method)
	assert.Equal(t, "spaceId=space-1", c.Query)
	assert.Equal(t, testAPIKey, c.Headers.Get("X-Service-Api-Key"))
	assert.Equal(t, testRequestOrgID, c.Headers.Get("X-Organization-Id"))
	require.Len(t, out, 1)
	assert.Equal(t, "wb-1", out[0].WorkbookID)
}

func Test_QuestionAPIClient_CreateWorkbook_shouldSendJsonBody(t *testing.T) {
	t.Parallel()
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	id, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookRequest{
		SpaceID:     "space-1",
		Title:       "T",
		Description: "D",
		Visibility:  "public",
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, "wb-new", id)
	require.Len(t, *calls, 1)
	c := (*calls)[0]
	assert.Equal(t, http.MethodPost, c.Method)
	assert.Equal(t, "application/json", c.Headers.Get("Content-Type"))
	var got seed.CreateWorkbookRequest
	require.NoError(t, json.Unmarshal(c.Body, &got))
	assert.Equal(t, "space-1", got.SpaceID)
	assert.Equal(t, "public", got.Visibility)
}

func Test_QuestionAPIClient_AddQuestion_shouldHitWorkbookScopedPath(t *testing.T) {
	t.Parallel()
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionRequest{
		QuestionType: "word_fill",
		Content:      "hello",
		Tags:         []string{"seed:wb-1:q-1"},
		OrderIndex:   0,
	})

	// then
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	c := (*calls)[0]
	assert.Equal(t, "/api/v1/internal/workbook/wb-1/question", c.Path)
}

func Test_QuestionAPIClient_shouldReturnError_whenServerReturnsNon2xx(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"code":"forbidden"}`)
	}))
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.ListWorkbooks(context.Background(), testRequestOrgID, "space-1")

	// then
	require.ErrorContains(t, err, "status 403")
}
