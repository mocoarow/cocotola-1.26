//go:build small

package gateway_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

const (
	testAPIKey       = "test-api-key"
	testRequestOrgID = "org-1"
	// listQuestionsLimitBytes mirrors the unexported limit in question_api_client.go.
	listQuestionsLimitBytes = 32 << 20
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
			_, _ = w.Write([]byte(`{"workbooks":[{"workbookId":"wb-1","title":"Existing","description":"desc [seedKey:k0]","visibility":"public","language":"ja","seedKey":"k1"}]}`))
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"workbookId":"wb-new","seedKey":"k-new"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/internal/workbook/wb-1", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, recordedCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    body,
		})
		mu.Unlock()

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"workbookId":"wb-1","seedKey":"k-updated"}`))
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

func newClient(srv *httptest.Server) *gateway.QuestionAPIClient {
	return gateway.NewQuestionAPIClient(srv.URL, testAPIKey, srv.Client())
}

func newStatusServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
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

func Test_QuestionAPIClient_ListWorkbooks_shouldReturnEveryFieldOfResponse(t *testing.T) {
	t.Parallel()

	// given
	srv, _ := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	out, err := client.ListWorkbooks(context.Background(), testRequestOrgID, "space-1")

	// then
	require.NoError(t, err)
	assert.Equal(t, []seed.WorkbookListItem{
		{WorkbookID: "wb-1", Title: "Existing", Description: "desc [seedKey:k0]", Visibility: "public", Language: "ja", SeedKey: "k1"},
	}, out)
}

func Test_QuestionAPIClient_CreateWorkbook_shouldSendJsonBody(t *testing.T) {
	t.Parallel()
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	out, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookInput{
		SpaceID:     "space-1",
		Title:       "T",
		Description: "D",
		Language:    "en",
	})

	// then
	require.NoError(t, err)
	assert.Equal(t, "wb-new", out.WorkbookID)
	require.Len(t, *calls, 1)
	c := (*calls)[0]
	assert.Equal(t, http.MethodPost, c.Method)
	assert.Equal(t, "application/json", c.Headers.Get("Content-Type"))
	assert.JSONEq(t, `{"spaceId":"space-1","title":"T","description":"D","visibility":"public","language":"en"}`, string(c.Body))
}

func Test_QuestionAPIClient_AddQuestion_shouldSendInputAsJsonBody(t *testing.T) {
	t.Parallel()

	// given
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{
		QuestionType: "word_fill",
		Content:      "hello",
		Tags:         []string{"seed-wb-1:q-1", "level:b1"},
		OrderIndex:   3,
	})

	// then
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	assert.JSONEq(t, `{"questionType":"word_fill","content":"hello","tags":["seed-wb-1:q-1","level:b1"],"orderIndex":3}`, string((*calls)[0].Body))
}

func Test_QuestionAPIClient_AddQuestion_shouldHitWorkbookScopedPath(t *testing.T) {
	t.Parallel()
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{
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

func Test_QuestionAPIClient_ListQuestions_shouldDecodeAllQuestions_whenResponseExceedsOneMiB(t *testing.T) {
	t.Parallel()

	// given: enough questions that the JSON body is larger than 1 MiB
	const questionCount = 20000
	questions := make([]map[string]any, questionCount)
	for i := range questions {
		questions[i] = map[string]any{
			"questionId": fmt.Sprintf("01900000-0000-7000-8000-%012d", i),
			"tags":       []string{fmt.Sprintf("seed-cefr-b1-wordfill-v1:%d", i), "level:b1", "source:tatoeba"},
		}
	}
	body, err := json.Marshal(map[string]any{"questions": questions})
	require.NoError(t, err)
	require.Greater(t, len(body), 1<<20)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	client := newClient(srv)

	// when
	got, err := client.ListQuestions(context.Background(), testRequestOrgID, "wb-1")

	// then
	require.NoError(t, err)
	assert.Len(t, got, questionCount)
}

func Test_QuestionAPIClient_ListQuestions_shouldReturnEveryFieldOfResponse(t *testing.T) {
	t.Parallel()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"questions":[{"questionId":"q-1","tags":["seed-wb:q-1","level:b1"]}]}`)
	}))
	defer srv.Close()
	client := newClient(srv)

	// when
	got, err := client.ListQuestions(context.Background(), testRequestOrgID, "wb-1")

	// then
	require.NoError(t, err)
	assert.Equal(t, []seed.QuestionListItem{
		{QuestionID: "q-1", Tags: []string{"seed-wb:q-1", "level:b1"}},
	}, got)
}

// newOversizedListQuestionsServer returns a server whose ListQuestions body is
// larger than listQuestionsLimitBytes: a single question whose ID alone fills it.
func newOversizedListQuestionsServer(t *testing.T) *httptest.Server {
	t.Helper()
	body := `{"questions":[{"questionId":"` + strings.Repeat("a", listQuestionsLimitBytes) + `","tags":[]}]}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
}

func Test_QuestionAPIClient_ListQuestions_shouldReturnMaxBytesError_whenResponseExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	srv := newOversizedListQuestionsServer(t)
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.ListQuestions(context.Background(), testRequestOrgID, "wb-1")

	// then
	var maxErr *http.MaxBytesError
	require.ErrorAs(t, err, &maxErr)
	assert.Equal(t, int64(listQuestionsLimitBytes), maxErr.Limit)
}

func Test_QuestionAPIClient_ListQuestions_shouldReportLimitInMessage_whenResponseExceedsLimit(t *testing.T) {
	t.Parallel()

	// given
	srv := newOversizedListQuestionsServer(t)
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.ListQuestions(context.Background(), testRequestOrgID, "wb-1")

	// then
	require.ErrorContains(t, err, fmt.Sprintf("response body exceeds %d bytes", listQuestionsLimitBytes))
}

func Test_QuestionAPIClient_AddQuestion_shouldKeepHTTPStatusError_whenServerReturns400(t *testing.T) {
	t.Parallel()

	// given
	srv := newStatusServer(http.StatusBadRequest)
	defer srv.Close()
	client := newClient(srv)

	// when
	err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{})

	// then
	var statusErr *gateway.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusBadRequest, statusErr.StatusCode)
}

func Test_QuestionAPIClient_AddQuestion_shouldReturnErrQuestionRejected_whenServerReturns400(t *testing.T) {
	t.Parallel()

	// given
	srv := newStatusServer(http.StatusBadRequest)
	defer srv.Close()
	client := newClient(srv)

	// when
	err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{})

	// then
	require.ErrorIs(t, err, seed.ErrQuestionRejected)
}

func Test_QuestionAPIClient_AddQuestion_shouldNotReturnErrQuestionRejected_whenServerReturnsNon400(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
	}{
		{name: "forbidden", status: http.StatusForbidden},
		{name: "internal server error", status: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			srv := newStatusServer(tt.status)
			defer srv.Close()
			client := newClient(srv)

			// when
			err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{})

			// then
			assert.NotErrorIs(t, err, seed.ErrQuestionRejected)
		})
	}
}

func Test_QuestionAPIClient_AddQuestion_shouldReturnHTTPStatusError_whenServerReturnsNon400(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
	}{
		{name: "forbidden", status: http.StatusForbidden},
		{name: "internal server error", status: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			srv := newStatusServer(tt.status)
			defer srv.Close()
			client := newClient(srv)

			// when
			err := client.AddQuestion(context.Background(), testRequestOrgID, "wb-1", seed.AddQuestionInput{})

			// then
			var statusErr *gateway.HTTPStatusError
			require.ErrorAs(t, err, &statusErr)
			assert.Equal(t, tt.status, statusErr.StatusCode)
		})
	}
}

func Test_QuestionAPIClient_CreateWorkbook_shouldReturnHTTPStatusError_whenServerReturns400(t *testing.T) {
	t.Parallel()

	// given
	srv := newStatusServer(http.StatusBadRequest)
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookInput{})

	// then
	var statusErr *gateway.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusBadRequest, statusErr.StatusCode)
}

func Test_QuestionAPIClient_CreateWorkbook_shouldNotReturnErrQuestionRejected_whenServerReturns400(t *testing.T) {
	t.Parallel()

	// given
	srv := newStatusServer(http.StatusBadRequest)
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookInput{})

	// then
	assert.NotErrorIs(t, err, seed.ErrQuestionRejected)
}

func Test_QuestionAPIClient_shouldIncludeStatusInMessage_whenServerReturnsNon2xx(t *testing.T) {
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

func Test_QuestionAPIClient_ListWorkbooks_shouldReturnErrSeedKeyNotStored_whenWorkbookHasNoSeedKeyField(t *testing.T) {
	t.Parallel()

	// given: a question service that predates the seedKey field
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"workbooks":[{"workbookId":"wb-1","title":"Existing","description":"desc","visibility":"public","language":"ja"}]}`))
	}))
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.ListWorkbooks(context.Background(), testRequestOrgID, "space-1")

	// then
	require.ErrorIs(t, err, seed.ErrSeedKeyNotStored)
}

func Test_QuestionAPIClient_ListWorkbooks_shouldReturnEmptySeedKey_whenSeedKeyFieldIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"workbooks":[{"workbookId":"wb-1","title":"Existing","description":"desc","visibility":"public","language":"ja","seedKey":""}]}`))
	}))
	defer srv.Close()
	client := newClient(srv)

	// when
	out, err := client.ListWorkbooks(context.Background(), testRequestOrgID, "space-1")

	// then
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Empty(t, out[0].SeedKey)
}

func Test_QuestionAPIClient_CreateWorkbook_shouldSendSeedKey_whenInputHasSeedKey(t *testing.T) {
	t.Parallel()

	// given
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookInput{
		SpaceID:     "space-1",
		Title:       "T",
		Description: "D",
		Language:    "en",
		SeedKey:     "k-new",
	})

	// then
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	assert.JSONEq(t, `{"spaceId":"space-1","title":"T","description":"D","visibility":"public","language":"en","seedKey":"k-new"}`, string((*calls)[0].Body))
}

func Test_QuestionAPIClient_CreateWorkbook_shouldReturnStoredSeedKey(t *testing.T) {
	t.Parallel()

	// given
	srv, _ := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	out, err := client.CreateWorkbook(context.Background(), testRequestOrgID, seed.CreateWorkbookInput{SpaceID: "space-1", Title: "T", Language: "en", SeedKey: "k-sent"})

	// then
	require.NoError(t, err)
	assert.Equal(t, "k-new", out.SeedKey)
}

func Test_QuestionAPIClient_UpdateWorkbook_shouldSendInputAsJsonBody(t *testing.T) {
	t.Parallel()

	// given
	srv, calls := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.UpdateWorkbook(context.Background(), testRequestOrgID, "wb-1", seed.UpdateWorkbookInput{
		Title:       "T",
		Description: "D",
		Visibility:  "public",
		Language:    "ja",
		SeedKey:     "k1",
	})

	// then
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	c := (*calls)[0]
	assert.Equal(t, http.MethodPut, c.Method)
	assert.Equal(t, testAPIKey, c.Headers.Get("X-Service-Api-Key"))
	assert.Equal(t, testRequestOrgID, c.Headers.Get("X-Organization-Id"))
	assert.JSONEq(t, `{"title":"T","description":"D","visibility":"public","language":"ja","seedKey":"k1"}`, string(c.Body))
}

func Test_QuestionAPIClient_UpdateWorkbook_shouldReturnStoredSeedKey(t *testing.T) {
	t.Parallel()

	// given
	srv, _ := newRecorder()
	defer srv.Close()
	client := newClient(srv)

	// when
	got, err := client.UpdateWorkbook(context.Background(), testRequestOrgID, "wb-1", seed.UpdateWorkbookInput{Title: "T", Visibility: "public", Language: "ja", SeedKey: "k1"})

	// then
	require.NoError(t, err)
	assert.Equal(t, "k-updated", got)
}

func Test_QuestionAPIClient_UpdateWorkbook_shouldReturnHTTPStatusError_whenServerReturnsNon2xx(t *testing.T) {
	t.Parallel()

	// given
	srv := newStatusServer(http.StatusConflict)
	defer srv.Close()
	client := newClient(srv)

	// when
	_, err := client.UpdateWorkbook(context.Background(), testRequestOrgID, "wb-1", seed.UpdateWorkbookInput{})

	// then
	var statusErr *gateway.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusConflict, statusErr.StatusCode)
}
