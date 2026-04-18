package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// GetAppUserIDFromContext extracts a domain.AppUserID from the Gin context or
// returns a zero value + false if not set. Middleware stores the user ID as a string.
func GetAppUserIDFromContext(c *gin.Context) (domain.AppUserID, bool) {
	s := c.GetString(controller.ContextFieldUserID{})
	if s == "" {
		return domain.AppUserID{}, false
	}
	id, err := domain.ParseAppUserID(s)
	if err != nil {
		slog.Warn("parse user ID from context", slog.String("raw", s), slog.Any("error", err))
		return domain.AppUserID{}, false
	}
	return id, !id.IsZero()
}
