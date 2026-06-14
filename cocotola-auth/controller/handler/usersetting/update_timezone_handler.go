package usersetting

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// errTimezoneDisallowedChars is returned by the bind step when the request
// timezone contains characters outside timezonePattern's allow-list. Kept
// as a sentinel so structured logs surface this distinct failure mode
// instead of conflating it with json-decode errors.
var errTimezoneDisallowedChars = errors.New("timezone contains characters outside the allowed set")

// timezonePattern enforces the OpenAPI character-class contract at bind
// time. oapi-codegen does not translate JSON-Schema `pattern` into
// validator/v10 tags, so without this guard the only bind-time check would
// be `required,max=64`; control characters, query-string punctuation, and
// other unexpected bytes would otherwise be passed all the way down to
// `time.LoadLocation` before being rejected. Matching here keeps the wire
// contract honest as a defense-in-depth measure (the runtime check still
// catches well-formed-but-nonexistent zones like "Not/AZone" later).
var timezonePattern = regexp.MustCompile(`^[A-Za-z_/+\-0-9]+$`)

// UpdateTimezoneHandler exposes the authenticated user's preferred IANA
// timezone as a mutable resource. The struct owns the persistence
// dependency and a slog logger tagged with its handler name; the actual
// bind / load / validate / save flow is delegated to updateRunner so this
// type only declares the field-specific request type, the character-class
// guard, and the Change call.
type UpdateTimezoneHandler struct {
	updateRunner
}

// NewUpdateTimezoneHandler returns a new UpdateTimezoneHandler.
func NewUpdateTimezoneHandler(settingSaver userSettingFinderSaver) *UpdateTimezoneHandler {
	return &UpdateTimezoneHandler{newUpdateRunner("UpdateTimezoneHandler", settingSaver)}
}

// UpdateTimezone handles PUT /auth/user-setting/timezone. The
// authenticated user's IANA timezone is replaced with the value in the
// request body. If no UserSetting row exists for the user, a default one
// is created with the requested timezone applied. Validation runs in two
// layers: the timezonePattern character-class guard at bind time, then
// time.LoadLocation in the domain to enforce that the name actually
// exists in the runtime tz database. Responses: 204 on success, 400 for
// any malformed/disallowed/unresolvable value, 401 if the caller is not
// authenticated, 409 on a concurrent modification, 500 on persistence
// failure.
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
