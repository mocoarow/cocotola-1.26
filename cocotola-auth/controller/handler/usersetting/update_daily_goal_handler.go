package usersetting

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// UpdateDailyGoalHandler updates the authenticated user's daily goal.
type UpdateDailyGoalHandler struct {
	updateRunner
}

// NewUpdateDailyGoalHandler returns a new UpdateDailyGoalHandler.
func NewUpdateDailyGoalHandler(settingSaver userSettingFinderSaver) *UpdateDailyGoalHandler {
	return &UpdateDailyGoalHandler{newUpdateRunner("UpdateDailyGoalHandler", settingSaver)}
}

// UpdateDailyGoal handles PUT /auth/user-setting/daily-goal.
func (h *UpdateDailyGoalHandler) UpdateDailyGoal(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context) (changeFn, error) {
			var req api.UpdateUserDailyGoalRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update daily goal request: %w", err)
			}

			return func(s *domain.UserSetting) error { return s.ChangeDailyGoal(req.DailyGoal) }, nil
		},
		invalidBindLog:   "invalid update daily goal request",
		invalidChangeLog: "change daily goal",
		invalidChangeMsg: "daily goal is invalid",
	})
}
