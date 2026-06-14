package usersetting

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// UpdateDailyGoalHandler exposes the authenticated user's daily problem
// count goal as a mutable resource. The struct owns the persistence
// dependency and a slog logger tagged with its handler name; the actual
// bind / load / validate / save flow is delegated to updateRunner so this
// type only declares the field-specific request type and Change call.
type UpdateDailyGoalHandler struct {
	updateRunner
}

// NewUpdateDailyGoalHandler returns a new UpdateDailyGoalHandler.
func NewUpdateDailyGoalHandler(settingSaver userSettingFinderSaver) *UpdateDailyGoalHandler {
	return &UpdateDailyGoalHandler{newUpdateRunner("UpdateDailyGoalHandler", settingSaver)}
}

// UpdateDailyGoal handles PUT /auth/user-setting/daily-goal. The
// authenticated user's daily problem goal is replaced with the value in
// the request body. If no UserSetting row exists for the user, a default
// one is created with the requested goal applied. Responses: 204 on
// success, 400 for an invalid request body (malformed JSON or
// out-of-range goal), 401 if the caller is not authenticated, 409 on a
// concurrent modification, 500 on persistence failure.
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
