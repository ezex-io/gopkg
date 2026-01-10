package scheduler

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"golang.org/x/sync/errgroup"
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
func (b EveryBuilder) Do(callback func(context.Context)) {
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
					callback(b.ctx)
				}()
			}
		}
	}()
}

type Scheduler struct {
	ctx  context.Context
	jobs []Job
	name string
}

func NewScheduler(ctx context.Context, name string) Scheduler {
	return Scheduler{
		ctx:  ctx,
		jobs: make([]Job, 0),
		name: name,
	}
}

func (s *Scheduler) AddJob(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start starts the scheduler and runs the jobs on the given interval.
func (s *Scheduler) Start(interval time.Duration, onSuccess func()) {
	Every(s.ctx, interval).Do(func(ctx context.Context) {
		s.runJobs(ctx, onSuccess)
	})
}

func (s *Scheduler) runJobs(ctx context.Context, onSuccess func()) {
	group, ctx := errgroup.WithContext(ctx)

	for _, j := range s.jobs {
		job := j
		group.Go(func() error {
			if err := job.Run(ctx); err != nil {
				log.Printf("job failed: %v", err)

				return err
			}

			return nil
		})
	}

	if err := group.Wait(); err == nil && onSuccess != nil {
		onSuccess()
	}
}
