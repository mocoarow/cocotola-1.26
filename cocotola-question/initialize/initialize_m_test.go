//go:build medium

package initialize_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/initialize"
)

const testAPIKey = "test-secret-key-that-is-long-enough"

func setupQuestionRouter(t *testing.T) *gin.Engine {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST not set; skipping Firestore integration test")
	}

	r := gin.New()
	cleanup, err := initialize.Initialize(
		context.Background(),
		r,
		config.QuestionConfig{
			FirestoreProjectID: "test-project",
			APIKey:             testAPIKey,
			AudioPublicBaseURL: "https://storage.example.com/audio",
		},
		gateway.NewAuthMiddleware("http://127.0.0.1:0", http.DefaultClient),
		middleware.NewAPIKeyMiddleware(testAPIKey),
		nil, nil, nil, nil, nil,
	)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	return r
}

func serveInternal(t *testing.T, r *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, path, nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", testAPIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func Test_Initialize_shouldServePendingAudio_whenOrganizationIDHeaderIsMissing(t *testing.T) {
	t.Parallel()

	// given
	r := setupQuestionRouter(t)

	// when
	w := serveInternal(t, r, http.MethodGet, "/internal/audio/questions/pending")

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_Initialize_shouldRejectInternalWorkbook_whenOrganizationIDHeaderIsMissing(t *testing.T) {
	t.Parallel()

	// given
	r := setupQuestionRouter(t)

	// when
	w := serveInternal(t, r, http.MethodPost, "/internal/workbook")

	// then
	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "missing_organization_id", resp.Code)
}
