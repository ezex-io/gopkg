package retry

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteAsync_Success(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	successCalled := make(chan bool, 1)
	failureCalled := make(chan bool, 1)

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		successCalled <- true

		return nil
	}, func(error) {
		failureCalled <- true
	})

	// Wait for success callback
	select {
	case <-successCalled:
		// Success
	case <-failureCalled:
		t.Error("Expected success callback, but failure callback was called")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Expected function to be called once")
}

func TestExecuteAsync_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	expectedCalls := int32(2)
	successCalled := make(chan bool, 1)
	failureCalled := make(chan bool, 1)

	ExecuteAsync(ctx, func() error {
		count := atomic.AddInt32(&callCount, 1)
		if count < expectedCalls {
			return errors.New("temporary error")
		}
		successCalled <- true

		return nil
	}, func(error) {
		failureCalled <- true
	}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(10*time.Millisecond))

	// Wait for success callback
	select {
	case <-successCalled:
		// Success
	case <-failureCalled:
		t.Error("Expected success callback, but failure callback was called")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, expectedCalls, atomic.LoadInt32(&callCount),
		"Expected function to be called %d times", expectedCalls)
}

func TestExecuteAsync_AllRetriesFail(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	expectedError := errors.New("persistent error")
	maxRetries := 3
	successCalled := make(chan bool, 1)
	failureCalled := make(chan error, 1)

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)

		return expectedError
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(maxRetries), WithAsyncRetryDelay(10*time.Millisecond))

	// Wait for failure callback
	select {
	case err := <-failureCalled:
		assert.Equal(t, expectedError, err)
	case <-successCalled:
		t.Error("Expected failure callback, but success callback was called")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(maxRetries), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_ContextCancellationDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := int32(0)
	failureCalled := make(chan error, 1)

	ExecuteAsync(ctx, func() error {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			// Cancel after first attempt
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
		}

		return errors.New("temporary error")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(5), WithAsyncRetryDelay(100*time.Millisecond))

	// Wait for failure callback
	select {
	case err := <-failureCalled:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	// Should be called at least once, but not all 5 times
	count := atomic.LoadInt32(&callCount)
	assert.GreaterOrEqual(t, count, int32(1), "Should be called at least once")
	assert.Less(t, count, int32(5), "Should not complete all retries due to cancellation")
}

func TestExecuteAsync_NilCallbacks(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	done := make(chan bool, 1)

	// Test with nil callbacks - should not panic
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		done <- true

		return nil
	}, nil)

	// Wait for function to complete
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for function to complete")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 10
	var wg sync.WaitGroup
	successCount := int32(0)
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			ExecuteAsync(ctx, func() error {
				atomic.AddInt32(&successCount, 1)
				done <- true

				return nil
			}, func(error) {
				t.Error("Failure callback should not be called")
			})

			select {
			case <-done:
			case <-time.After(1 * time.Second):
				t.Error("Timeout waiting for callback")
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&successCount))
}
