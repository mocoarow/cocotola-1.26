// Package study provides HTTP handlers for spaced repetition study operations.
package study

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// InitStudyRouter sets up the routes for study operations under the given parent router group.
func InitStudyRouter(
	getStudyQuestionsHandler *GetStudyQuestionsHandler,
	recordAnswerHandler *RecordAnswerHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	studyGroup := parentRouterGroup.Group("workbook/:workbookId/study")
	studyGroup.Use(authMiddleware)
	studyGroup.Use(middleware...)

	studyGroup.GET("", getStudyQuestionsHandler.GetStudyQuestions)
	studyGroup.POST("/:questionId/answer", recordAnswerHandler.RecordAnswer)
}

func handleStudyError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	if errors.Is(err, domain.ErrForbidden) {
		logger.WarnContext(ctx, "forbidden", slog.Any("error", err))
		c.JSON(http.StatusForbidden, controller.NewErrorResponse("forbidden", http.StatusText(http.StatusForbidden)))
		return
	}
	if errors.Is(err, domain.ErrWorkbookNotFound) {
		logger.WarnContext(ctx, "workbook not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("workbook_not_found", "workbook not found"))
		return
	}
	if errors.Is(err, domain.ErrQuestionNotFound) {
		logger.WarnContext(ctx, "question not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("question_not_found", "question not found"))
		return
	}
	if errors.Is(err, domain.ErrStudyRecordNotFound) {
		logger.WarnContext(ctx, "study record not found", slog.Any("error", err))
		c.JSON(http.StatusNotFound, controller.NewErrorResponse("study_record_not_found", "study record not found"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
