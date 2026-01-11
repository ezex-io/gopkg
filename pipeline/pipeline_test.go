package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClosePipeline(t *testing.T) {
	pipe := New[int](t.Context())

	pipe.RegisterReceiver(func(a int) {
		time.Sleep(time.Duration(a) * time.Millisecond)
	})

	for i := 0; i < 10; i++ {
		go pipe.Send(i)
	}

	pipe.Close()
	pipe.Close()

	assert.Eventually(t, pipe.IsClosed, 500*time.Millisecond, 10*time.Millisecond)
}

func TestSendReceive(t *testing.T) {
	pipe := New[float64](t.Context())

	received := make(chan float64, 1)
	receiver := func(data float64) {
		received <- data
	}

	pipe.RegisterReceiver(receiver)

	// Test multiple messages
	pipe.Send(3.141592)
	pipe.Send(2.718281)
	pipe.Send(1.618033)

	assert.InEpsilon(t, 3.141592, <-received, 0.0)
	assert.InEpsilon(t, 2.718281, <-received, 0.0)
	assert.InEpsilon(t, 1.618033, <-received, 0.0)
}

func TestMultipleReceivers(t *testing.T) {
	pipe := New[int](t.Context())

	first := make(chan int, 1)
	second := make(chan int, 1)

	pipe.RegisterReceiver(func(v int) { first <- v })
	pipe.RegisterReceiver(func(v int) { second <- v })

	pipe.Send(42)

	select {
	case val := <-first:
		assert.Equal(t, 42, val)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("first receiver did not receive value")
	}

	select {
	case val := <-second:
		assert.Equal(t, 42, val)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second receiver did not receive value")
	}
}

// TestSendAfterClose verifies error handling.
func TestSendAfterClose(t *testing.T) {
	pipe := New[string](t.Context())

	// Close the pipeline first
	pipe.Close()

	// Send should fail gracefully
	assert.NotPanics(t, func() {
		pipe.Send("should not panic")
	})
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	pipe := New[string](ctx)

	// Cancel the context
	cancel()

	// Send should fail gracefully
	assert.NotPanics(t, func() {
		pipe.Send("should not panic after cancel")
	})
}

func TestClose(t *testing.T) {
	pipe := New[int](t.Context())

	assert.False(t, pipe.IsClosed())
	pipe.Close()
	assert.True(t, pipe.IsClosed())

	// Second close should be no-op
	pipe.Close()
	assert.True(t, pipe.IsClosed())
}

func TestDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 0)
	defer cancel()

	pipe := New[int](ctx)

	receiverCalled := make(chan struct{})
	pipe.RegisterReceiver(func(_ int) {
		close(receiverCalled)
	})

	// Wait for context to timeout
	time.Sleep(100 * time.Millisecond)

	pipe.Send(42)

	// Verify receiver wasn't called
	select {
	case <-receiverCalled:
		t.Fatal("receiver should not be called for failed sends")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message should be received
	}

	// Verify pipeline is still operational for other cases
	assert.False(t, pipe.IsClosed(), "pipeline should not be closed just because send failed")
}

func TestUnsafeGetChannel(t *testing.T) {
	pipe := New[int](t.Context())

	pipeCh := pipe.UnsafeGetChannel()
	assert.NotNil(t, pipeCh)

	testValue := 123
	pipe.Send(testValue)

	select {
	case val := <-pipeCh:
		assert.Equal(t, testValue, val)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for channel value")
	}
}

func TestOptionsConfigurePipeline(t *testing.T) {
	name := "custom"
	buf := 7

	pipe := New[int](t.Context(), WithName(name), WithBufferSize(buf))

	pipeCh := pipe.UnsafeGetChannel()
	assert.NotNil(t, pipeCh)

	assert.Equal(t, name, pipe.Name())
	assert.Equal(t, buf, cap(pipeCh))
}
