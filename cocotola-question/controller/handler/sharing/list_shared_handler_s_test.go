package sharing_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

func Test_ListSharedHandler_ListShared_shouldReturnWorkbookDetails_whenUsecaseReturnsItems(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	addedAt := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	usecases := newSharingUsecases(t)
	usecases.listShared.On("ListShared", mock.Anything, mock.Anything).
		Return(&referenceservice.ListSharedOutput{References: []referenceservice.SharedItem{{
			ReferenceID: "ref-1",
			WorkbookID:  "wb-1",
			Title:       "TOEIC 600",
			Description: "Vocabulary for TOEIC 600",
			Language:    "en",
			AddedAt:     addedAt,
		}}}, nil).Once()
	r := initSharingRouter(ctx, t, usecases, fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

	// when
	w := serve(ctx, t, r, http.MethodGet, "/api/v1/workbook/shared")
	respBytes := readBytes(t, w.Body)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListSharedResponse
	require.NoError(t, json.Unmarshal(respBytes, &resp))
	assert.Equal(t, []api.SharedItemResponse{{
		ReferenceID: "ref-1",
		WorkbookID:  "wb-1",
		Title:       "TOEIC 600",
		Description: "Vocabulary for TOEIC 600",
		Language:    "en",
		AddedAt:     addedAt,
	}}, resp.References)
}
