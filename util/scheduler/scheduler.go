package scheduler

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

func After(ctx context.Context, duration time.Duration, callback func()) {
	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			callback()
		}
	}()
}

func Every(ctx context.Context, duration time.Duration, callback func()) {
	go func() {
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
					callback()
				}()
			}
		}
	}()
}

type IScheduler interface {
	Submit(job Job)
	Run()
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

func (s *Scheduler) Submit(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start starts the scheduler and runs the jobs on the given interval.
func (s *Scheduler) Start(interval time.Duration, onSuccess func()) {
	Every(s.ctx, interval, func() {
		s.runJobs(onSuccess)
	})
}

func (s *Scheduler) runJobs(onSuccess func()) {
	var wg sync.WaitGroup
	var success atomic.Bool
	success.Store(true)

	for _, j := range s.jobs {
		job := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := job.Run(s.ctx); err != nil {
				log.Printf("job failed: %v", err)
				success.Store(false)
			}
		}()
	}

	wg.Wait()
	if success.Load() && onSuccess != nil {
		onSuccess()
	}
}
