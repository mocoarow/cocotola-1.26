package gateway

import (
	"context"
	"time"
)

func newShutdownCtx(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), timeout)
}
