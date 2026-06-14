package study

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"

	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// defaultDashboardDays is the contribution-graph window applied when the
// caller does not supply a `days` query parameter. It is a handler-level
// convention (HTTP default) rather than a domain invariant, so it lives
// here instead of in studyservice — unlike the min/max bounds which come
// from the service-layer constants below.
const defaultDashboardDays = 365

// GetDashboardUsecase defines the use case method required by the
// GetDashboardHandler.
type GetDashboardUsecase interface {
	GetDashboard(ctx context.Context, input *studyservice.GetDashboardInput) (*studyservice.GetDashboardOutput, error)
}

// GetDashboardHandler handles GET /api/v1/study/dashboard. The endpoint is
// user-scoped (not workbook-scoped) so the dashboard aggregates activity
// across every workbook the operator studies.
type GetDashboardHandler struct {
	usecase GetDashboardUsecase
	logger  *slog.Logger
}

// NewGetDashboardHandler returns a new GetDashboardHandler.
func NewGetDashboardHandler(usecase GetDashboardUsecase) *GetDashboardHandler {
	return &GetDashboardHandler{
		usecase: usecase,
		logger:  slog.Default().With(slog.String(liblogging.LoggerNameKey, "GetDashboardHandler")),
	}
}

// GetDashboard handles GET /api/v1/study/dashboard?days=<n>.
func (h *GetDashboardHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetString(controller.ContextFieldUserID{})
	if userID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid user ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	organizationID := c.GetString(controller.ContextFieldOrganizationID{})
	if organizationID == "" {
		h.logger.WarnContext(ctx, "unauthorized: missing or invalid organization ID")
		c.JSON(http.StatusUnauthorized, controller.NewErrorResponse("unauthorized", http.StatusText(http.StatusUnauthorized)))
		return
	}

	days, ok := h.parseDays(ctx, c)
	if !ok {
		return
	}

	// Field-level pre-checks so the 400 body tells the client exactly
	// which input failed. The downstream NewGetDashboardInput validator
	// would otherwise funnel every cause through the same generic
	// "invalid request parameters" string. X-Local-Timezone is read by
	// /workbook/.../answer (the write side persists it on each bucket)
	// but is intentionally not consumed here: the dashboard window is
	// computed purely from the user-local YYYY-MM-DD that the frontend
	// already resolved before sending X-Local-Date.
	todayKey := c.GetHeader("X-Local-Date")
	if todayKey == "" {
		h.logger.WarnContext(ctx, "missing X-Local-Date header")
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "X-Local-Date header is required"))
		return
	}

	input, err := studyservice.NewGetDashboardInput(userID, organizationID, days, todayKey)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid dashboard input", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "invalid request parameters"))
		return
	}

	output, err := h.usecase.GetDashboard(ctx, input)
	if err != nil {
		handleStudyError(ctx, h.logger, c, "get dashboard", err)
		return
	}

	items := make([]api.DashboardDailyItem, 0, len(output.Days))
	for _, d := range output.Days {
		items = append(items, api.DashboardDailyItem{
			Date:          d.Date,
			AnsweredCount: int32(d.AnsweredCount),
			CorrectCount:  int32(d.CorrectCount),
		})
	}

	c.JSON(http.StatusOK, api.GetDashboardResponse{
		From:          output.From,
		To:            output.To,
		Days:          items,
		CurrentStreak: int32(output.CurrentStreak),
		LongestStreak: int32(output.LongestStreak),
		TodayCount:    int32(output.TodayCount),
		TodayCorrect:  int32(output.TodayCorrect),
		ActiveDays:    int32(output.ActiveDays),
		TotalAnswered: int32(output.TotalAnswered),
		TotalCorrect:  int32(output.TotalCorrect),
	})
}

func (h *GetDashboardHandler) parseDays(ctx context.Context, c *gin.Context) (int, bool) {
	daysParam := c.Query("days")
	if daysParam == "" {
		return defaultDashboardDays, true
	}
	days, err := strconv.Atoi(daysParam)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid days parameter", slog.String("days", daysParam), slog.Any("error", err))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", "days must be an integer"))
		return 0, false
	}
	if days < studyservice.MinDashboardDays || days > studyservice.MaxDashboardDays {
		h.logger.WarnContext(ctx, "days out of range", slog.Int("days", days))
		c.JSON(http.StatusBadRequest, controller.NewErrorResponse("invalid_request", fmt.Sprintf("days must be between %d and %d", studyservice.MinDashboardDays, studyservice.MaxDashboardDays)))
		return 0, false
	}
	return days, true
}

// InitDashboardRouter sets up the user-scoped study-dashboard route. The
// route lives at /api/v1/study/dashboard (not under /workbook/:id) because
// the dashboard aggregates across every workbook.
func InitDashboardRouter(
	getDashboardHandler *GetDashboardHandler,
	parentRouterGroup gin.IRouter,
	authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	dashboardGroup := parentRouterGroup.Group("study")
	dashboardGroup.Use(authMiddleware)
	dashboardGroup.Use(middleware...)

	dashboardGroup.GET("/dashboard", getDashboardHandler.GetDashboard)
}
