// Package study provides HTTP handlers for spaced repetition study operations.
package study

import (
	"github.com/gin-gonic/gin"
)

// InitStudyRouter sets up the routes for study operations under the given parent router group.
func InitStudyRouter(
	getStudyQuestionsHandler *GetStudyQuestionsHandler,
	getStudySummaryHandler *GetStudySummaryHandler,
	recordAnswerHandler *RecordAnswerHandler,
	deleteStudyHistoryHandler *DeleteStudyHistoryHandler,
	listStudyRecordsHandler *ListStudyRecordsHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	studyGroup := parentRouterGroup.Group("workbook/:workbookId/study")
	studyGroup.Use(authMiddleware)
	studyGroup.Use(middleware...)

	studyGroup.GET("", getStudyQuestionsHandler.GetStudyQuestions)
	studyGroup.GET("/summary", getStudySummaryHandler.GetStudySummary)
	studyGroup.GET("/records", listStudyRecordsHandler.ListStudyRecords)
	studyGroup.POST("/:questionId/answer", recordAnswerHandler.RecordAnswer)
	studyGroup.DELETE("", deleteStudyHistoryHandler.DeleteStudyHistory)
}
