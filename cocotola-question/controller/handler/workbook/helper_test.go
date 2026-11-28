package workbook_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	workbookhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/workbook"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

const (
	fixtureUserID         = "user-1"
	fixtureOrganizationID = "org-1"
	fixtureWorkbookID     = "wb-1"
	fixtureSpaceID        = "space-1"
	workbookPath          = "/api/v1/workbook"
)

type workbookUsecases struct {
	get    *MockGetWorkbookUsecase
	list   *MockListWorkbooksUsecase
	delete *MockDeleteWorkbookUsecase
}

func newWorkbookUsecases(t *testing.T) workbookUsecases {
	t.Helper()

	return workbookUsecases{
		get:    NewMockGetWorkbookUsecase(t),
		list:   NewMockListWorkbooksUsecase(t),
		delete: NewMockDeleteWorkbookUsecase(t),
	}
}

func initWorkbookRouter(ctx context.Context, t *testing.T, userID string, organizationID string, usecases workbookUsecases) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)

	workbookhandler.InitWorkbookRouter(
		workbookhandler.NewCreateWorkbookHandler(NewMockCreateWorkbookUsecase(t)),
		workbookhandler.NewGetWorkbookHandler(usecases.get),
		workbookhandler.NewListWorkbooksHandler(usecases.list),
		workbookhandler.NewUpdateWorkbookHandler(NewMockUpdateWorkbookUsecase(t)),
		workbookhandler.NewDeleteWorkbookHandler(usecases.delete),
		router.Group("api").Group("v1"),
		fakeOperatorMiddleware(userID, organizationID),
	)

	return router
}

func serve(ctx context.Context, t *testing.T, r *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	return w
}

func fixtureItem(suffix string) workbookservice.Item {
	return workbookservice.Item{
		WorkbookID:     "wb-" + suffix,
		SpaceID:        "space-" + suffix,
		OwnerID:        "owner-" + suffix,
		OrganizationID: "org-" + suffix,
		Title:          "title-" + suffix,
		Description:    "description-" + suffix,
		Visibility:     "private",
		Language:       "fr",
		SeedKey:        "seed-" + suffix,
		CreatedAt:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC),
	}
}

func toWorkbookResponse(item workbookservice.Item) api.WorkbookResponse {
	return api.WorkbookResponse{
		WorkbookID:     item.WorkbookID,
		SpaceID:        item.SpaceID,
		OwnerID:        item.OwnerID,
		OrganizationID: item.OrganizationID,
		Title:          item.Title,
		Description:    item.Description,
		Visibility:     api.WorkbookResponseVisibility(item.Visibility),
		Language:       item.Language,
		SeedKey:        item.SeedKey,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

var serverConfig libcontroller.ServerConfig

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	serverConfig = libcontroller.ServerConfig{
		CORS: libcontroller.CORSConfig{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Content-Type,Authorization,X-Service-Api-Key",
		},
		Log: libcontroller.LogConfig{
			AccessLog:             false,
			AccessLogRequestBody:  false,
			AccessLogResponseBody: false,
		},
		Debug: libcontroller.DebugConfig{
			Gin:  false,
			Wait: false,
		},
	}
	os.Exit(m.Run())
}

func fakeOperatorMiddleware(userID string, organizationID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldOrganizationID{}, organizationID)
		c.Next()
	}
}

func initInternalWorkbookRouter(ctx context.Context, t *testing.T, createUsecase *MockCreateWorkbookUsecase, updateUsecase *MockUpdateWorkbookUsecase) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	internal := router.Group("api").Group("v1").Group("internal")
	internal.Use(fakeOperatorMiddleware(fixtureUserID, fixtureOrganizationID))

	workbookhandler.InitInternalWorkbookRouter(
		workbookhandler.NewCreateWorkbookHandler(createUsecase),
		workbookhandler.NewListWorkbooksHandler(NewMockListWorkbooksUsecase(t)),
		workbookhandler.NewUpdateWorkbookHandler(updateUsecase),
		workbookhandler.NewDeleteWorkbookHandler(NewMockDeleteWorkbookUsecase(t)),
		internal,
	)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func validateErrorResponse(t *testing.T, respBytes []byte, expectedErrorCode string, expectedErrorMessage string) {
	t.Helper()

	jsonObj, err := oj.Parse(respBytes)
	require.NoError(t, err)

	codeExpr, err := jp.ParseString("$.code")
	require.NoError(t, err)
	code := codeExpr.Get(jsonObj)
	require.Len(t, code, 1, "response should have one code: %+v", jsonObj)
	assert.Equal(t, expectedErrorCode, code[0])

	messageExpr, err := jp.ParseString("$.message")
	require.NoError(t, err)
	message := messageExpr.Get(jsonObj)
	require.Len(t, message, 1, "response should have one message: %+v", jsonObj)
	assert.Equal(t, expectedErrorMessage, message[0])
}
