package contextWithTimer

import (
	"context"
	"time"
)

func New(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc, func()) {
	ctx, cancel := context.WithCancel(ctx)
	timer := time.NewTimer(duration)

	go func() {
		select {
		case <-timer.C:
		case <-ctx.Done():
		}

		timer.Stop()
		cancel()
		return
	}()

	return ctx, cancel, func() { timer.Reset(duration) }
}
