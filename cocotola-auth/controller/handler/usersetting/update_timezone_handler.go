package usersetting

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var errTimezoneDisallowedChars = errors.New("timezone contains characters outside the allowed set")

// timezonePattern mirrors the OpenAPI `pattern`, which oapi-codegen does not emit as a validator tag.
var timezonePattern = regexp.MustCompile(`^[A-Za-z_/+\-0-9]+$`)

// UpdateTimezoneHandler updates the authenticated user's IANA timezone.
type UpdateTimezoneHandler struct {
	updateRunner
}

// NewUpdateTimezoneHandler returns a new UpdateTimezoneHandler.
func NewUpdateTimezoneHandler(settingSaver userSettingFinderSaver) *UpdateTimezoneHandler {
	return &UpdateTimezoneHandler{newUpdateRunner("UpdateTimezoneHandler", settingSaver)}
}

// UpdateTimezone handles PUT /auth/user-setting/timezone.
func (h *UpdateTimezoneHandler) UpdateTimezone(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context) (changeFn, error) {
			var req api.UpdateUserTimezoneRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update timezone request: %w", err)
			}

			if !timezonePattern.MatchString(req.Timezone) {
				return nil, errTimezoneDisallowedChars
			}

			return func(s *domain.UserSetting) error { return s.ChangeTimezone(req.Timezone) }, nil
		},
		invalidBindLog:   "invalid update timezone request",
		invalidChangeLog: "change timezone",
		invalidChangeMsg: "timezone is invalid",
	})
}
