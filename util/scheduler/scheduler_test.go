package scheduler

import (
	"bytes"
	"context"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestAfterRunsWhenNotCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan struct{})
	go After(ctx, 5*time.Millisecond, func() {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for After to run")
	}
}

func TestAfterStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var called atomic.Bool
	go After(ctx, 20*time.Millisecond, func() {
		called.Store(true)
	})

	cancel()
	time.Sleep(30 * time.Millisecond)

	if called.Load() {
		t.Fatal("After callback should not run after cancellation")
	}
}

func TestEveryRunsUntilContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var count atomic.Int32

	done := make(chan struct{})
	go func() {
		Every(ctx, 2*time.Millisecond, func() {
			if count.Add(1) == 3 {
				cancel()
			}
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for Every to stop after cancellation")
	}

	if got := count.Load(); got != 3 {
		t.Fatalf("expected 3 executions before cancel, got %d", got)
	}
}

func TestEveryRecoversFromPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var count atomic.Int32

	// Silence panic log noise while still validating recovery path.
	origOutput := log.Writer()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() {
		log.SetOutput(origOutput)
	})

	done := make(chan struct{})
	go func() {
		Every(ctx, 2*time.Millisecond, func() {
			n := count.Add(1)
			if n == 1 {
				panic("boom")
			}
			if n >= 2 {
				cancel()
			}
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for Every to continue after panic")
	}

	if got := count.Load(); got < 2 {
		t.Fatalf("expected at least 2 executions despite panic, got %d", got)
	}
	if !bytes.Contains(buf.Bytes(), []byte("panic in job")) {
		t.Fatal("expected panic to be logged")
	}
}
