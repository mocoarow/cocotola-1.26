package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// NewWaitMiddleware blocks the request pipeline for the given duration while
// still honoring the request context so cancellation propagates immediately.
func NewWaitMiddleware(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if duration <= 0 {
			c.Next()
			return
		}

		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-c.Request.Context().Done():
		}

		c.Next()
	}
}
