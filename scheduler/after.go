package scheduler

import (
	"context"
	"time"
)

type AfterBuilder struct {
	ctx      context.Context
	duration time.Duration
}

// After schedules a one-time execution after the given duration.
func After(ctx context.Context, duration time.Duration) AfterBuilder {
	return AfterBuilder{ctx: ctx, duration: duration}
}

// Do registers the callback to run once after the configured delay.
// The scheduler passes the builder's context to the callback for cancellation-aware work.
func (b AfterBuilder) Do(callback func(context.Context)) {
	go func() {
		timer := time.NewTimer(b.duration)
		defer timer.Stop()

		select {
		case <-b.ctx.Done():
			return
		case <-timer.C:
			callback(b.ctx)
		}
	}()
}
