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

	firestoreClient, shutdownFirestore, err := gateway.InitFirestore(context.Background(), "test-project")
	require.NoError(t, err)
	t.Cleanup(shutdownFirestore)

	r := gin.New()
	initialize.Initialize(r, config.QuestionConfig{
		FirestoreProjectID: "test-project",
		APIKey:             testAPIKey,
		AudioPublicBaseURL: "https://storage.example.com/audio",
	}, firestoreClient, initialize.Dependencies{
		AuthMiddleware:      middleware.NewAuthMiddleware(gateway.NewAuthServiceMeFetcher("http://127.0.0.1:0", http.DefaultClient, 0)),
		APIKeyMiddleware:    middleware.NewAPIKeyMiddleware(testAPIKey),
		AuthzChecker:        nil,
		OrgResolver:         nil,
		MaxWorkbooksFetcher: nil,
		SpaceTypeFetcher:    nil,
		PolicyAdder:         nil,
	})

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

func servePublic(t *testing.T, r *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, path, nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func Test_Initialize_shouldServeHealthCheck_whenUnauthenticated(t *testing.T) {
	t.Parallel()

	// given
	r := setupQuestionRouter(t)

	// when
	w := servePublic(t, r, http.MethodGet, "/question/health")

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_Initialize_shouldRejectRequest_whenPublicRouteIsCalledWithoutCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "workbook", method: http.MethodGet, path: "/workbook"},
		{name: "question", method: http.MethodGet, path: "/workbook/wb-1/question"},
		{name: "sharing", method: http.MethodPost, path: "/workbook/wb-1/share"},
		{name: "study", method: http.MethodGet, path: "/workbook/wb-1/study"},
		{name: "dashboard", method: http.MethodGet, path: "/study/dashboard"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			r := setupQuestionRouter(t)

			// when
			w := servePublic(t, r, tt.method, tt.path)

			// then
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
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
