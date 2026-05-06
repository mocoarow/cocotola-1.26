package study_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/study"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	fixtureUserID         = "user-1"
	fixtureOrganizationID = "org-1"
	fixtureWorkbookID     = "wb-1"
	fixtureQuestionID     = "q-1"
)

func fakeAuthMiddleware(userID string, organizationName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldOrganizationName{}, organizationName)
		c.Next()
	}
}

func fakeOrgResolverMiddleware(orgID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldOrganizationID{}, orgID)
		c.Next()
	}
}

func noopMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func initStudyRouter(ctx context.Context, t *testing.T, getUsecase *MockGetStudyQuestionsUsecase, recordUsecase *MockRecordAnswerUsecase) *gin.Engine {
	t.Helper()
	return initStudyRouterWithMiddleware(ctx, t, getUsecase, NewMockGetStudySummaryUsecase(t), recordUsecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
}

func initStudyRouterWithSummary(ctx context.Context, t *testing.T, summaryUsecase *MockGetStudySummaryUsecase) *gin.Engine {
	t.Helper()
	return initStudyRouterWithMiddleware(ctx, t, NewMockGetStudyQuestionsUsecase(t), summaryUsecase, NewMockRecordAnswerUsecase(t), fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
}

func initStudyRouterWithMiddleware(ctx context.Context, t *testing.T, getUsecase *MockGetStudyQuestionsUsecase, summaryUsecase *MockGetStudySummaryUsecase, recordUsecase *MockRecordAnswerUsecase, authMiddleware gin.HandlerFunc, orgMiddleware gin.HandlerFunc) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	api := router.Group("api")
	v1 := api.Group("v1")

	getStudyQuestionsHandler := studyhandler.NewGetStudyQuestionsHandler(getUsecase)
	getStudySummaryHandler := studyhandler.NewGetStudySummaryHandler(summaryUsecase)
	recordAnswerHandler := studyhandler.NewRecordAnswerHandler(recordUsecase)
	studyhandler.InitStudyRouter(getStudyQuestionsHandler, getStudySummaryHandler, recordAnswerHandler, v1, authMiddleware, orgMiddleware)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func parseJSON(t *testing.T, data []byte) any {
	t.Helper()
	obj, err := oj.Parse(data)
	require.NoError(t, err)
	return obj
}

func parseExpr(t *testing.T, v string) jp.Expr {
	t.Helper()
	expr, err := jp.ParseString(v)
	require.NoError(t, err)
	return expr
}

func validateErrorResponse(t *testing.T, respBytes []byte, expectedErrorCode string, expectedErrorMessage string) {
	t.Helper()

	jsonObj := parseJSON(t, respBytes)

	errorCodeExpr := parseExpr(t, "$.code")
	errorCode := errorCodeExpr.Get(jsonObj)
	require.Len(t, errorCode, 1, "response should have one code: %+v", jsonObj)
	assert.Equal(t, expectedErrorCode, errorCode[0])

	errorMessageExpr := parseExpr(t, "$.message")
	errorMessage := errorMessageExpr.Get(jsonObj)
	require.Len(t, errorMessage, 1, "response should have one message: %+v", jsonObj)
	assert.Equal(t, expectedErrorMessage, errorMessage[0])
}
