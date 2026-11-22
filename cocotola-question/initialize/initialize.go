// Package initialize provides a reusable initialization function for the cocotola-question module.
package initialize

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
	healthhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/health"
	questionhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/question"
	sharinghandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/sharing"
	studyhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/study"
	workbookhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/workbook"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/gateway"
	questionusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/question"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
	studyusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/study"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"
)

// AuthorizationChecker checks if an action is allowed by RBAC policy.
type AuthorizationChecker interface {
	IsAllowed(ctx context.Context, organizationID string, operatorID string, action domain.Action, resource domain.Resource) (bool, error)
}

// MaxWorkbooksFetcher fetches the max workbooks limit for a user.
type MaxWorkbooksFetcher interface {
	FetchMaxWorkbooks(ctx context.Context, userID string) (int, error)
}

// SpaceTypeFetcher resolves a space's type via cocotola-auth.
type SpaceTypeFetcher interface {
	FetchSpaceType(ctx context.Context, spaceID string) (domain.SpaceType, error)
}

// PolicyAdder adds per-user RBAC policies via the auth service.
type PolicyAdder interface {
	AddPolicyForUser(ctx context.Context, organizationID string, userID string, action domain.Action, resource domain.Resource, effect domain.Effect) error
}

// Dependencies holds the collaborators the question module is wired with.
type Dependencies struct {
	AuthMiddleware      gin.HandlerFunc
	APIKeyMiddleware    gin.HandlerFunc
	AuthzChecker        AuthorizationChecker
	OrgResolver         middleware.OrganizationIDResolver
	MaxWorkbooksFetcher MaxWorkbooksFetcher
	SpaceTypeFetcher    SpaceTypeFetcher
	PolicyAdder         PolicyAdder
}

// Initialize registers the question module's routes under parent.
func Initialize(parent gin.IRouter, questionConfig config.QuestionConfig, firestoreClient *firestore.Client, deps Dependencies) {
	repos := newRepositories(firestoreClient)
	usecases := newUsecases(repos, deps)
	initRoutes(parent, questionConfig, repos, usecases, deps)
}

type repositories struct {
	workbook           *gateway.WorkbookRepository
	question           *gateway.QuestionRepository
	reference          *gateway.ReferenceRepository
	ownedWorkbookList  *gateway.OwnedWorkbookListRepository
	studyRecord        *gateway.StudyRecordRepository
	studyDailyStats    *gateway.StudyDailyStatsRepository
	activeQuestionList *gateway.ActiveQuestionListRepository
	health             *gateway.HealthRepository
}

func newRepositories(client *firestore.Client) repositories {
	return repositories{
		workbook:           gateway.NewWorkbookRepository(client),
		question:           gateway.NewQuestionRepository(client),
		reference:          gateway.NewReferenceRepository(client),
		ownedWorkbookList:  gateway.NewOwnedWorkbookListRepository(client),
		studyRecord:        gateway.NewStudyRecordRepository(client),
		studyDailyStats:    gateway.NewStudyDailyStatsRepository(client),
		activeQuestionList: gateway.NewActiveQuestionListRepository(client),
		health:             gateway.NewHealthRepository(client),
	}
}

type usecases struct {
	workbook   *workbookusecase.Command
	question   *questionusecase.Command
	audioBatch *questionusecase.AudioBatchCommand
	sharing    *sharingusecase.Command
	study      *studyusecase.Command
}

func newUsecases(r repositories, deps Dependencies) usecases {
	contentParser := contentjson.Parser{}
	return usecases{
		workbook: workbookusecase.NewCommand(r.workbook, r.workbook, r.workbook, r.ownedWorkbookList, r.ownedWorkbookList, deps.MaxWorkbooksFetcher, deps.SpaceTypeFetcher, deps.AuthzChecker, deps.PolicyAdder),
		question: questionusecase.NewCommand(r.question, r.question, r.question, r.question, r.workbook, deps.AuthzChecker, r.activeQuestionList, r.activeQuestionList, contentParser),
		audioBatch: questionusecase.NewAudioBatchCommand(r.question, r.question, r.question, r.question, r.question, contentParser, questionusecase.UsecaseConfig{
			ClockFunc: nil,
		}),
		sharing: sharingusecase.NewCommand(r.reference, r.reference, r.reference, r.workbook, r.workbook, r.workbook, deps.AuthzChecker),
		study: studyusecase.NewCommand(r.studyRecord, r.activeQuestionList, r.question, r.workbook, deps.AuthzChecker, r.studyDailyStats, contentParser, studyusecase.UsecaseConfig{
			ClockFunc:   nil,
			ShuffleFunc: nil,
		}),
	}
}

type routeGroups struct {
	public            gin.IRouter
	internal          gin.IRouter
	internalOrg       gin.IRouter
	auth              gin.HandlerFunc
	organizationScope gin.HandlerFunc
}

func initRoutes(parent gin.IRouter, questionConfig config.QuestionConfig, r repositories, uc usecases, deps Dependencies) {
	internal := parent.Group("internal", deps.APIKeyMiddleware)
	groups := routeGroups{
		public:            parent,
		internal:          internal,
		internalOrg:       internal.Group("", middleware.NewOrganizationIDHeaderMiddleware()),
		auth:              deps.AuthMiddleware,
		organizationScope: middleware.NewOrganizationResolverMiddleware(deps.OrgResolver),
	}
	audioURLBuilder := handler.NewAudioURLBuilder(questionConfig.AudioPublicBaseURL)

	healthhandler.InitRouter(healthhandler.NewCheckHandler(r.health), groups.public)
	initWorkbookRoutes(groups, uc.workbook)
	initQuestionRoutes(groups, uc.question, uc.audioBatch, audioURLBuilder)
	initSharingRoutes(groups, uc.sharing)
	initStudyRoutes(groups, uc.study, audioURLBuilder)
}

func initWorkbookRoutes(g routeGroups, uc *workbookusecase.Command) {
	create := workbookhandler.NewCreateWorkbookHandler(uc)
	get := workbookhandler.NewGetWorkbookHandler(uc)
	list := workbookhandler.NewListWorkbooksHandler(uc)
	update := workbookhandler.NewUpdateWorkbookHandler(uc)
	del := workbookhandler.NewDeleteWorkbookHandler(uc)
	workbookhandler.InitWorkbookRouter(create, get, list, update, del, g.public, g.auth, g.organizationScope)
	workbookhandler.InitInternalWorkbookRouter(create, list, update, del, g.internalOrg)
}

func initQuestionRoutes(g routeGroups, uc *questionusecase.Command, audioUC *questionusecase.AudioBatchCommand, audioURLBuilder *handler.AudioURLBuilder) {
	response := questionhandler.NewResponseBuilder(audioURLBuilder)
	add := questionhandler.NewAddQuestionHandler(uc, response)
	get := questionhandler.NewGetQuestionHandler(uc, response)
	list := questionhandler.NewListQuestionsHandler(uc, response)
	update := questionhandler.NewUpdateQuestionHandler(uc, response)
	del := questionhandler.NewDeleteQuestionHandler(uc)
	questionhandler.InitQuestionRouter(add, get, list, update, del, g.public, g.auth, g.organizationScope)
	questionhandler.InitInternalQuestionRouter(add, list, update, del, g.internalOrg)
	// The audio queue is processed across organizations, so it does not require X-Organization-Id.
	questionhandler.InitInternalAudioRouter(questionhandler.NewAudioHandler(audioUC), g.internal)
}

func initSharingRoutes(g routeGroups, uc *sharingusecase.Command) {
	sharinghandler.InitSharingRouter(
		sharinghandler.NewShareWorkbookHandler(uc),
		sharinghandler.NewListSharedHandler(uc),
		sharinghandler.NewUnshareHandler(uc),
		sharinghandler.NewListPublicHandler(uc),
		g.public, g.auth, g.organizationScope)
}

func initStudyRoutes(g routeGroups, uc *studyusecase.Command, audioURLBuilder *handler.AudioURLBuilder) {
	studyhandler.InitStudyRouter(
		studyhandler.NewGetStudyQuestionsHandler(uc, audioURLBuilder),
		studyhandler.NewGetStudySummaryHandler(uc),
		studyhandler.NewRecordAnswerHandler(uc),
		studyhandler.NewDeleteStudyHistoryHandler(uc),
		studyhandler.NewListStudyRecordsHandler(uc),
		g.public, g.auth, g.organizationScope)
	studyhandler.InitDashboardRouter(studyhandler.NewGetDashboardHandler(uc), g.public, g.auth, g.organizationScope)
}
