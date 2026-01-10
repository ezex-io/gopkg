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
	runs *atomic.Int32
}

func (j testJob) Run(_ context.Context) error {
	j.runs.Add(1)

	return nil
}

type errorJob struct {
	runs   *atomic.Int32
	cancel context.CancelFunc
}

func (j errorJob) Run(_ context.Context) error {
	j.runs.Add(1)
	j.cancel()

	return errors.New("job failed")
}

func TestStartWithOnSuccessOption(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var success atomic.Int32
	var runs atomic.Int32

	s := scheduler.NewScheduler(ctx, "test-scheduler")
	s.AddJob(testJob{runs: &runs})

	s.Start(1*time.Millisecond, scheduler.WithOnSuccess(func() {
		success.Add(1)
		cancel()
	}))

	select {
	case <-ctx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for onSuccess to be called")
	}

	if runs.Load() == 0 {
		t.Fatal("expected job to run at least once")
	}
	if success.Load() == 0 {
		t.Fatal("expected onSuccess to be invoked")
	}
}

func TestOnSuccessNotCalledWhenJobErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var success atomic.Int32
	var runs atomic.Int32

	s := scheduler.NewScheduler(ctx, "test-scheduler")
	s.AddJob(errorJob{runs: &runs, cancel: cancel})

	s.Start(1*time.Millisecond, scheduler.WithOnSuccess(func() {
		success.Add(1)
	}))

	select {
	case <-ctx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for job error to cancel context")
	}

	if runs.Load() == 0 {
		t.Fatal("expected job to run at least once")
	}
	if success.Load() != 0 {
		t.Fatal("onSuccess should not be invoked when a job errors")
	}
}
