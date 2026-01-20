package scheduler

import (
	"context"
	"log"
	"runtime/debug"
	"time"
)

type EveryBuilder struct {
	duration time.Duration
}

// Every schedules a callback to run on the provided interval.
func Every(duration time.Duration) EveryBuilder {
	return EveryBuilder{duration: duration}
}

// Do registers the callback to run repeatedly on the configured interval.
// The scheduler passes the builder's context to the callback for cancellation-aware work.
func (b EveryBuilder) Do(ctx context.Context, callback func(ctx context.Context)) {
	go func() {
		ticker := time.NewTicker(b.duration)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
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
					callback(ctx)
				}()
			}
		}
	}()
}
