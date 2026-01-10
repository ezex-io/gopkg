package scheduler

import (
	"context"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

type Scheduler struct {
	ctx       context.Context
	jobs      []Job
	name      string
	onSuccess func()
}

type Option func(*Scheduler)

func NewScheduler(ctx context.Context, name string, opts ...Option) Scheduler {
	scheduler := Scheduler{
		ctx:  ctx,
		jobs: make([]Job, 0),
		name: name,
	}

	for _, opt := range opts {
		opt(&scheduler)
	}

	return scheduler
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
func (s *Scheduler) Start(interval time.Duration) {
	Every(s.ctx, interval).Do(func(ctx context.Context) {
		s.runJobs(ctx)
	})
}

func (s *Scheduler) runJobs(ctx context.Context) {
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

	err := group.Wait()
	if err == nil && s.onSuccess != nil {
		s.onSuccess()
	}
}
