package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/ezex-io/gopkg/scheduler"
)

func TestAfterNotCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	done := make(chan struct{})
	scheduler.After(ctx, 5*time.Millisecond).Do(func(_ context.Context) {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for After to run")
	}
}

func TestAfterCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	called := make(chan struct{})
	scheduler.After(ctx, 20*time.Millisecond).Do(func(_ context.Context) {
		close(called)
	})

	cancel()

	select {
	case <-ctx.Done():
	case <-called:
		t.Fatal("After callback should not run after cancellation")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for After to run")
	}
}
