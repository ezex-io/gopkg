package scheduler

import (
	"context"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

type Scheduler struct {
	jobs      []Job
	onSuccess func()
}

type Option func(*Scheduler)

func NewScheduler() Scheduler {
	return Scheduler{
		jobs: make([]Job, 0),
	}
}

// WithOnSuccess registers a callback to run after all jobs succeed in a tick.
func WithOnSuccess(cb func()) Option {
	return func(s *Scheduler) {
		s.onSuccess = cb
	}
}

func (s *Scheduler) AddJob(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start starts the scheduler and runs the jobs on the given interval.
func (s *Scheduler) Start(ctx context.Context, interval time.Duration, opts ...Option) {
	for _, opt := range opts {
		opt(s)
	}

	Every(interval).Do(ctx, func(ctx context.Context) {
		s.runJobs(ctx)
	})
}

func (s *Scheduler) runJobs(ctx context.Context) {
	group, _ := errgroup.WithContext(ctx)

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

	err := group.Wait()
	if err == nil && s.onSuccess != nil {
		s.onSuccess()
	}
}
