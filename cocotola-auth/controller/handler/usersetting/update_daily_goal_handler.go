package usersetting

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

// UpdateDailyGoalHandler updates the authenticated user's daily goal.
type UpdateDailyGoalHandler struct {
	updateRunner
}

// NewUpdateDailyGoalHandler returns a new UpdateDailyGoalHandler.
func NewUpdateDailyGoalHandler(updater userSettingUpdater) *UpdateDailyGoalHandler {
	return &UpdateDailyGoalHandler{newUpdateRunner("UpdateDailyGoalHandler", updater)}
}

// UpdateDailyGoal handles PUT /auth/user-setting/daily-goal.
func (h *UpdateDailyGoalHandler) UpdateDailyGoal(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context, appUserID domain.AppUserID) (applyFn, error) {
			var req api.UpdateUserDailyGoalRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update daily goal request: %w", err)
			}

			input, err := usersettingservice.NewUpdateDailyGoalInput(appUserID, req.DailyGoal)
			if err != nil {
				return nil, fmt.Errorf("new update daily goal input: %w", err)
			}

			return func(ctx context.Context) error { return h.updater.UpdateDailyGoal(ctx, input) }, nil
		},
		invalidBindLog:   "invalid update daily goal request",
		invalidChangeLog: "change daily goal",
		invalidChangeMsg: "daily goal is invalid",
	})
}
