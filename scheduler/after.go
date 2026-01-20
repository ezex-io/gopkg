package scheduler

import (
	"context"
	"time"
)

type AfterBuilder struct {
	duration time.Duration
}

// After schedules a one-time execution after the given duration.
func After(duration time.Duration) AfterBuilder {
	return AfterBuilder{duration: duration}
}

// Do registers the callback to run once after the configured delay.
// The scheduler passes the builder's context to the callback for cancellation-aware work.
func (b AfterBuilder) Do(ctx context.Context, callback func(ctx context.Context)) {
	go func(ctx context.Context) {
		timer := time.NewTimer(b.duration)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			callback(ctx)
		}
	}(ctx)
}
