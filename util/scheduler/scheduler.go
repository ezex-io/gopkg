package scheduler

import (
	"context"
	"log"
	"runtime/debug"
	"time"
)

func After(ctx context.Context, duration time.Duration, fn func()) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return

	case <-timer.C:
		fn()
	}
}

func Every(ctx context.Context, duration time.Duration, fn func()) {
	ticker := time.NewTicker(duration)
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
				fn()
			}()
		}
	}
}
