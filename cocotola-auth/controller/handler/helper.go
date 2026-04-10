package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// GetAppUserIDFromContext extracts a domain.AppUserID from the Gin context or
// returns a zero value + false if not set. Middleware stores the VO directly.
func GetAppUserIDFromContext(c *gin.Context) (domain.AppUserID, bool) {
	v, ok := c.Get(controller.ContextFieldUserID{})
	if !ok {
		return domain.AppUserID{}, false
	}
	id, ok := v.(domain.AppUserID)
	if !ok {
		return domain.AppUserID{}, false
	}
	return id, !id.IsZero()
}
