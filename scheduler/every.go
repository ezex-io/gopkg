package scheduler

import (
	"context"
	"log"
	"runtime/debug"
	"time"
)

type EveryBuilder struct {
	ctx      context.Context
	duration time.Duration
}

// Every schedules a callback to run on the provided interval.
func Every(ctx context.Context, duration time.Duration) EveryBuilder {
	return EveryBuilder{ctx: ctx, duration: duration}
}

// Do registers the callback to run repeatedly on the configured interval.
// The scheduler passes the builder's context to the callback for cancellation-aware work.
func (b EveryBuilder) Do(callback func()) {
	go func() {
		ticker := time.NewTicker(b.duration)
		defer ticker.Stop()

		for {
			select {
			case <-b.ctx.Done():
				return
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf(
								"scheduler: panic in job: %v\n%s",
								r,
								debug.Stack(),
							)
						}
					}()
					callback()
				}()
			}
		}
	}()
}
