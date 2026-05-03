// Package question provides HTTP handlers for question management operations.
package question

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
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
//
// The list endpoint is included so that batch jobs (e.g. cocotola-init seeding)
// can perform idempotency checks before adding questions.
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

func handleQuestionError(ctx context.Context, logger *slog.Logger, c *gin.Context, action string, err error) {
	if errors.Is(err, domain.ErrInvalidArgument) {
		logger.WarnContext(ctx, "invalid argument", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", http.StatusText(http.StatusBadRequest)))
		return
	}
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
	if errors.Is(err, domain.ErrConcurrentModification) {
		logger.WarnContext(ctx, "concurrent modification", slog.Any("error", err))
		c.JSON(http.StatusConflict, controller.NewErrorResponse("conflict", "resource was modified concurrently"))
		return
	}
	logger.ErrorContext(ctx, action, slog.Any("error", err))
	c.JSON(http.StatusInternalServerError, controller.NewErrorResponse("internal_server_error", http.StatusText(http.StatusInternalServerError)))
}
