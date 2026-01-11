package scheduler_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/ezex-io/gopkg/scheduler"
)

func TestEveryNotCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	done := make(chan struct{})
	count := 0
	scheduler.Every(ctx, 2*time.Millisecond).Do(func() {
		count++
		if count == 3 {
			close(done)
		}
	})

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Every to run")
	}

	if count != 3 {
		t.Fatalf("expected 3 executions, got %d", count)
	}
}

func TestEveryCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	called := make(chan struct{})
	scheduler.Every(ctx, 20*time.Millisecond).Do(func() {
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

func TestEveryRecoversFromPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	// Silence panic log noise while still validating recovery path.
	origOutput := log.Writer()
	var buf bytes.Buffer
	log.SetOutput(&buf)

	t.Cleanup(func() {
		log.SetOutput(origOutput)
	})

	done := make(chan struct{})
	count := 0
	scheduler.Every(ctx, 2*time.Millisecond).Do(func() {
		count++
		if count == 1 {
			panic("boom")
		}
		if count >= 2 {
			cancel()
			close(done)
		}
	})

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for Every to continue after panic")
	}

	if count < 2 {
		t.Fatalf("expected at least 2 executions despite panic, got %d", count)
	}
	if !bytes.Contains(buf.Bytes(), []byte("panic in job")) {
		t.Fatal("expected panic to be logged")
	}
}
