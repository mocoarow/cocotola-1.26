package question_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func invalidTagsTests() []struct {
	name string
	tags []string
} {
	tooMany := make([]string, 21)
	for i := range tooMany {
		tooMany[i] = "tag"
	}
	return []struct {
		name string
		tags []string
	}{
		{name: "too many tags", tags: tooMany},
		{name: "too long tag", tags: []string{strings.Repeat("a", 101)}},
	}
}

func questionBody(t *testing.T, tags []string) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"questionType": "word_card",
		"content":      "content",
		"tags":         tags,
		"orderIndex":   0,
	})
	require.NoError(t, err)
	return string(body)
}

func Test_AddQuestionHandler_AddQuestion_shouldReturn400_whenTagsAreInvalid(t *testing.T) {
	t.Parallel()

	for _, tt := range invalidTagsTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			r := initInternalQuestionRouter(ctx, t, NewMockAddQuestionUsecase(t), NewMockUpdateQuestionUsecase(t))
			w := httptest.NewRecorder()

			// when
			url := "/api/v1/internal/workbook/" + fixtureWorkbookID + "/question"
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(questionBody(t, tt.tags)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, readBytes(t, w.Body), "invalid_request", http.StatusText(http.StatusBadRequest))
		})
	}
}

func Test_UpdateQuestionHandler_UpdateQuestion_shouldReturn400_whenTagsAreInvalid(t *testing.T) {
	t.Parallel()

	for _, tt := range invalidTagsTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			r := initInternalQuestionRouter(ctx, t, NewMockAddQuestionUsecase(t), NewMockUpdateQuestionUsecase(t))
			w := httptest.NewRecorder()

			// when
			url := "/api/v1/internal/workbook/" + fixtureWorkbookID + "/question/" + fixtureQuestionID
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(questionBody(t, tt.tags)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, readBytes(t, w.Body), "invalid_request", http.StatusText(http.StatusBadRequest))
		})
	}
}
