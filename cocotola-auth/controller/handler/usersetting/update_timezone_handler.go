package usersetting

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

var errTimezoneDisallowedChars = errors.New("timezone contains characters outside the allowed set")

// timezonePattern mirrors the OpenAPI `pattern`, which oapi-codegen does not emit as a validator tag.
var timezonePattern = regexp.MustCompile(`^[A-Za-z_/+\-0-9]+$`)

// UpdateTimezoneHandler updates the authenticated user's IANA timezone.
type UpdateTimezoneHandler struct {
	updateRunner
}

// NewUpdateTimezoneHandler returns a new UpdateTimezoneHandler.
func NewUpdateTimezoneHandler(updater userSettingUpdater) *UpdateTimezoneHandler {
	return &UpdateTimezoneHandler{newUpdateRunner("UpdateTimezoneHandler", updater)}
}

// UpdateTimezone handles PUT /auth/user-setting/timezone.
func (h *UpdateTimezoneHandler) UpdateTimezone(c *gin.Context) {
	h.run(c, updateConfig{
		bindRequest: func(c *gin.Context, appUserID domain.AppUserID) (applyFn, error) {
			var req api.UpdateUserTimezoneRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, fmt.Errorf("decode update timezone request: %w", err)
			}

			if !timezonePattern.MatchString(req.Timezone) {
				return nil, errTimezoneDisallowedChars
			}

			input, err := usersettingservice.NewUpdateTimezoneInput(appUserID, req.Timezone)
			if err != nil {
				return nil, fmt.Errorf("new update timezone input: %w", err)
			}

			return func(ctx context.Context) error { return h.updater.UpdateTimezone(ctx, input) }, nil
		},
		invalidBindLog:   "invalid update timezone request",
		invalidChangeLog: "change timezone",
		invalidChangeMsg: "timezone is invalid",
	})
}
