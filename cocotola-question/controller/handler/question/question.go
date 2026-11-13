// Package question provides HTTP handlers for question management operations.
package question

import (
	"github.com/gin-gonic/gin"
)

// InitQuestionRouter sets up the routes for question operations under the given parent router group.
func InitQuestionRouter(
	addHandler *AddQuestionHandler,
	getHandler *GetQuestionHandler,
	listHandler *ListQuestionsHandler,
	updateHandler *UpdateQuestionHandler,
	deleteHandler *DeleteQuestionHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	questionGroup := parentRouterGroup.Group("workbook/:workbookId/question")
	questionGroup.Use(authMiddleware)
	questionGroup.Use(middleware...)

	questionGroup.POST("", addHandler.AddQuestion)
	questionGroup.GET("", listHandler.ListQuestions)
	questionGroup.GET("/:questionId", getHandler.GetQuestion)
	questionGroup.PUT("/:questionId", updateHandler.UpdateQuestion)
	questionGroup.DELETE("/:questionId", deleteHandler.DeleteQuestion)
}

// InitInternalQuestionRouter mounts question CRUD endpoints on the internal
// parent router (protected by X-Service-Api-Key).
func InitInternalQuestionRouter(
	addHandler *AddQuestionHandler,
	listHandler *ListQuestionsHandler,
	updateHandler *UpdateQuestionHandler,
	deleteHandler *DeleteQuestionHandler,
	parentRouterGroup gin.IRouter,
) {
	questionGroup := parentRouterGroup.Group("workbook/:workbookId/question")
	questionGroup.POST("", addHandler.AddQuestion)
	questionGroup.GET("", listHandler.ListQuestions)
	questionGroup.PUT("/:questionId", updateHandler.UpdateQuestion)
	questionGroup.DELETE("/:questionId", deleteHandler.DeleteQuestion)
}
