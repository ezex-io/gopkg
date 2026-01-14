package scheduler_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ezex-io/gopkg/scheduler"
)

type testJob struct {
	counter *atomic.Int32
}

func (j testJob) Run() error {
	j.counter.Add(1)

	return nil
}

type errorJob struct {
	cancel context.CancelFunc
}

func (j errorJob) Run() error {
	j.cancel()

	return errors.New("job failed")
}

func TestSchedulerJobSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var counter atomic.Int32

	s := scheduler.NewScheduler(ctx)
	s.AddJob(testJob{counter: &counter})

	s.Start(1*time.Millisecond, scheduler.WithOnSuccess(func() {
		counter.Add(1)
		cancel()
	}))

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for onSuccess to be called")
	}

	if counter.Load() == 1 {
		t.Fatal("expected onSuccess to be invoked")
	}
}

func TestSchedulerJobError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	s := scheduler.NewScheduler(ctx)
	s.AddJob(errorJob{cancel: cancel})

	s.Start(1*time.Millisecond, scheduler.WithOnSuccess(func() {
		t.Fatal("onSuccess should not be invoked when a job errors")
	}))

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job error to cancel context")
	}
}

func TestStartMultipleJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var counter atomic.Int32

	s := scheduler.NewScheduler(ctx)
	s.AddJob(testJob{counter: &counter})
	s.AddJob(testJob{counter: &counter})
	s.AddJob(testJob{counter: &counter})

	s.Start(1*time.Millisecond,
		scheduler.WithOnSuccess(
			func() {
				cancel()
			},
		))

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for onSuccess to be called")
	}

	if counter.Load() != 3 {
		t.Fatalf("expected 3 executions, got %d", counter.Load())
	}
}
