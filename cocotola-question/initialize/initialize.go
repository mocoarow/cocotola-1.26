// Package initialize provides a reusable initialization function for the cocotola-question module.
package initialize

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	authcontroller "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	healthhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/health"
	questionhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/question"
	sharinghandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/sharing"
	workbookhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/workbook"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// OrganizationIDResolver resolves an organization name to its ID.
type OrganizationIDResolver func(ctx context.Context, name string) (int, error)

// AuthorizationChecker checks if an action is allowed by RBAC policy.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID int, operatorID int, action domainrbac.Action, resource domainrbac.Resource) (bool, error)
}

// Initialize sets up the cocotola-question module: gateway, usecase, and controller layers.
// It registers all question-related routes under the given parent router group and returns
// a cleanup function to close the Firestore client.
func Initialize(
	ctx context.Context,
	parent gin.IRouter,
	questionConfig config.QuestionConfig,
	authMiddleware gin.HandlerFunc,
	authzChecker AuthorizationChecker,
	orgResolver OrganizationIDResolver,
) (func(), error) {
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, "cocotola-question-init"))

	// gateway layer: Firestore client
	firestoreClient, err := gateway.NewFirestoreClient(ctx, questionConfig.FirestoreProjectID)
	if err != nil {
		return nil, fmt.Errorf("new firestore client: %w", err)
	}

	workbookRepo := gateway.NewWorkbookRepository(firestoreClient)
	questionRepo := gateway.NewQuestionRepository(firestoreClient)
	referenceRepo := gateway.NewReferenceRepository(firestoreClient)
	healthRepo := gateway.NewHealthRepository(firestoreClient)

	// organization resolver middleware
	orgResolverMiddleware := newOrganizationResolverMiddleware(orgResolver, logger)

	// usecase layer
	workbookCommand := workbookusecase.NewCommand(workbookRepo, workbookRepo, workbookRepo, workbookRepo, authzChecker)
	questionCommand := questionusecase.NewCommand(questionRepo, questionRepo, questionRepo, questionRepo, workbookRepo, authzChecker)
	sharingCommand := sharingusecase.NewCommand(referenceRepo, referenceRepo, referenceRepo, workbookRepo, workbookRepo, authzChecker)

	// controller layer
	checkHandler := healthhandler.NewCheckHandler(healthRepo)
	healthhandler.InitRouter(checkHandler, parent)

	createWorkbookHandler := workbookhandler.NewCreateWorkbookHandler(workbookCommand)
	getWorkbookHandler := workbookhandler.NewGetWorkbookHandler(workbookCommand)
	listWorkbooksHandler := workbookhandler.NewListWorkbooksHandler(workbookCommand)
	updateWorkbookHandler := workbookhandler.NewUpdateWorkbookHandler(workbookCommand)
	deleteWorkbookHandler := workbookhandler.NewDeleteWorkbookHandler(workbookCommand)
	workbookhandler.InitWorkbookRouter(createWorkbookHandler, getWorkbookHandler, listWorkbooksHandler, updateWorkbookHandler, deleteWorkbookHandler, parent, authMiddleware, orgResolverMiddleware)

	addQuestionHandler := questionhandler.NewAddQuestionHandler(questionCommand)
	getQuestionHandler := questionhandler.NewGetQuestionHandler(questionCommand)
	listQuestionsHandler := questionhandler.NewListQuestionsHandler(questionCommand)
	updateQuestionHandler := questionhandler.NewUpdateQuestionHandler(questionCommand)
	deleteQuestionHandler := questionhandler.NewDeleteQuestionHandler(questionCommand)
	questionhandler.InitQuestionRouter(addQuestionHandler, getQuestionHandler, listQuestionsHandler, updateQuestionHandler, deleteQuestionHandler, parent, authMiddleware, orgResolverMiddleware)

	shareWorkbookHandler := sharinghandler.NewShareWorkbookHandler(sharingCommand)
	listSharedHandler := sharinghandler.NewListSharedHandler(sharingCommand)
	unshareHandler := sharinghandler.NewUnshareHandler(sharingCommand)
	listPublicHandler := sharinghandler.NewListPublicHandler(sharingCommand)
	sharinghandler.InitSharingRouter(shareWorkbookHandler, listSharedHandler, unshareHandler, listPublicHandler, parent, authMiddleware, orgResolverMiddleware)

	cleanup := func() {
		if err := firestoreClient.Close(); err != nil {
			logger.Error("close firestore client", slog.Any("error", err))
		}
	}

	return cleanup, nil
}

func newOrganizationResolverMiddleware(resolver OrganizationIDResolver, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		orgName := c.GetString(authcontroller.ContextFieldOrganizationName{})
		if orgName == "" {
			logger.WarnContext(ctx, "missing organization name in context")
			c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
			c.Abort()
			return
		}

		orgID, err := resolver(ctx, orgName)
		if err != nil {
			logger.ErrorContext(ctx, "resolve organization ID", slog.String("organization_name", orgName), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
			c.Abort()
			return
		}

		c.Set(controller.ContextFieldOrganizationID{}, orgID)
		c.Next()
	}
}
