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
		return nil
	}, func() {
		successCalled <- true
	}, func(err error) {
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
		return nil
	}, func() {
		successCalled <- true
	}, func(err error) {
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

	assert.Equal(t, expectedCalls, atomic.LoadInt32(&callCount), "Expected function to be called %d times", expectedCalls)
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
	}, func() {
		successCalled <- true
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

func TestExecuteAsync_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := int32(0)
	successCalled := make(chan bool, 1)
	failureCalled := make(chan error, 1)

	// Cancel context immediately
	cancel()

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("should not succeed")
	}, func() {
		successCalled <- true
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(3))

	// Wait for failure callback
	select {
	case err := <-failureCalled:
		assert.ErrorIs(t, err, context.Canceled)
	case <-successCalled:
		t.Error("Expected failure callback, but success callback was called")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(0), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_ContextCancellationDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := int32(0)
	successCalled := make(chan bool, 1)
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
	}, func() {
		successCalled <- true
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(5), WithAsyncRetryDelay(100*time.Millisecond))

	// Wait for failure callback
	select {
	case err := <-failureCalled:
		assert.ErrorIs(t, err, context.Canceled)
	case <-successCalled:
		t.Error("Expected failure callback, but success callback was called")
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	// Should be called at least once, but not all 5 times
	count := atomic.LoadInt32(&callCount)
	assert.GreaterOrEqual(t, count, int32(1), "Should be called at least once")
	assert.Less(t, count, int32(5), "Should not complete all retries due to cancellation")
}

func TestExecuteAsync_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := int32(0)
	successCalled := make(chan bool, 1)
	failureCalled := make(chan error, 1)

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("temporary error")
	}, func() {
		successCalled <- true
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(10), WithAsyncRetryDelay(100*time.Millisecond))

	// Wait for failure callback
	select {
	case err := <-failureCalled:
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	case <-successCalled:
		t.Error("Expected failure callback, but success callback was called")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	// Should be called at least once, but not all 10 times due to timeout
	count := atomic.LoadInt32(&callCount)
	assert.GreaterOrEqual(t, count, int32(1), "Should be called at least once")
	assert.Less(t, count, int32(10), "Should not complete all retries due to timeout")
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
	}, nil, nil)

	// Wait for function to complete
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for function to complete")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_NilSuccessCallback(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	done := make(chan bool, 1)

	// Test with nil success callback
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		done <- true
		return nil
	}, nil, func(err error) {
		t.Error("Failure callback should not be called")
	})

	// Wait for function to complete
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for function to complete")
	}
}

func TestExecuteAsync_NilFailureCallback(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	done := make(chan bool, 1)

	// Test with nil failure callback
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		done <- true
		return errors.New("error")
	}, func() {
		t.Error("Success callback should not be called")
	}, nil, WithAsyncMaxRetries(1))

	// Wait for function to complete
	select {
	case <-done:
		// Function executed
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for function to complete")
	}
}

func TestExecuteAsync_CustomRetryCount(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	maxRetries := 5
	failureCalled := make(chan error, 1)

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("persistent error")
	}, func() {
		t.Error("Success callback should not be called")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(maxRetries), WithAsyncRetryDelay(1*time.Millisecond))

	// Wait for failure callback
	select {
	case <-failureCalled:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(maxRetries), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_CustomRetryDelay(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	retryDelay := 50 * time.Millisecond
	failureCalled := make(chan error, 1)

	start := time.Now()
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("persistent error")
	}, func() {
		t.Error("Success callback should not be called")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(retryDelay))

	// Wait for failure callback
	select {
	case <-failureCalled:
		elapsed := time.Since(start)
		// Should take at least 2 * retryDelay (2 delays between 3 attempts)
		expectedMinDuration := 2 * retryDelay
		assert.GreaterOrEqual(t, elapsed, expectedMinDuration, "Expected at least %v elapsed time", expectedMinDuration)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for callback")
	}
}

func TestExecuteAsync_CallbackCalledOnce(t *testing.T) {
	ctx := context.Background()
	successCount := int32(0)
	failureCount := int32(0)
	done := make(chan bool, 1)

	ExecuteAsync(ctx, func() error {
		return nil
	}, func() {
		atomic.AddInt32(&successCount, 1)
		// Small delay to ensure callback isn't called multiple times
		time.Sleep(50 * time.Millisecond)
		done <- true
	}, func(err error) {
		atomic.AddInt32(&failureCount, 1)
	})

	// Wait for success callback
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	// Verify callbacks were called exactly once
	assert.Equal(t, int32(1), atomic.LoadInt32(&successCount), "Expected success callback to be called once")
	assert.Equal(t, int32(0), atomic.LoadInt32(&failureCount), "Expected failure callback not to be called")
}

func TestExecuteAsync_NoRetryDelay(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	failureCalled := make(chan error, 1)

	start := time.Now()
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("persistent error")
	}, func() {
		t.Error("Success callback should not be called")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(0))

	// Wait for failure callback
	select {
	case <-failureCalled:
		elapsed := time.Since(start)
		// Should complete very quickly with no delay
		assert.Less(t, elapsed, 100*time.Millisecond, "Expected quick execution with no delay")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_SingleRetry(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	failureCalled := make(chan error, 1)

	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("persistent error")
	}, func() {
		t.Error("Success callback should not be called")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncMaxRetries(1), WithAsyncRetryDelay(1*time.Millisecond))

	// Wait for failure callback
	select {
	case <-failureCalled:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestExecuteAsync_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 10
	successCount := int32(0)
	done := make(chan bool, concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		ExecuteAsync(ctx, func() error {
			return nil
		}, func() {
			atomic.AddInt32(&successCount, 1)
			done <- true
		}, func(err error) {
			t.Error("Failure callback should not be called")
		})
	}

	// Wait for all callbacks
	for i := 0; i < concurrentCalls; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for callback")
		}
	}

	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&successCount))
}

func TestExecuteAsync_DefaultOptions(t *testing.T) {
	ctx := context.Background()
	callCount := int32(0)
	failureCalled := make(chan error, 1)

	// Use default syncOptions (3 retries, 2 second delay) but override delay to make test faster
	ExecuteAsync(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("persistent error")
	}, func() {
		t.Error("Success callback should not be called")
	}, func(err error) {
		failureCalled <- err
	}, WithAsyncRetryDelay(1*time.Millisecond))

	// Wait for failure callback
	select {
	case <-failureCalled:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for callback")
	}

	// Default is 3 retries
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount), "Expected function to be called 3 times (default)")
}

// Race condition tests

func TestExecuteAsync_RaceCondition_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 100
	var wg sync.WaitGroup
	successCount := int32(0)
	failureCount := int32(0)
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			ExecuteAsync(ctx, func() error {
				return nil
			}, func() {
				atomic.AddInt32(&successCount, 1)
				done <- true
			}, func(err error) {
				atomic.AddInt32(&failureCount, 1)
				done <- true
			}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(1*time.Millisecond))

			// Wait for callback
			select {
			case <-done:
			case <-time.After(1 * time.Second):
				t.Error("Timeout waiting for callback")
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&successCount), "All success callbacks should be called")
	assert.Equal(t, int32(0), atomic.LoadInt32(&failureCount), "No failure callbacks should be called")
}

func TestExecuteAsync_RaceCondition_SharedState(t *testing.T) {
	ctx := context.Background()
	var counter int32
	concurrentCalls := 50
	var wg sync.WaitGroup
	completedCount := int32(0)
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			ExecuteAsync(ctx, func() error {
				// Use atomic operations to properly synchronize access to shared state
				atomic.AddInt32(&counter, 1)
				return nil
			}, func() {
				atomic.AddInt32(&completedCount, 1)
				done <- true
			}, func(err error) {
				done <- true
			}, WithAsyncMaxRetries(1), WithAsyncRetryDelay(0))

			// Wait for callback
			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&completedCount), "All operations should complete")
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&counter), "All operations should increment counter")
}

func TestExecuteAsync_RaceCondition_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	concurrentCalls := 50
	var wg sync.WaitGroup
	callbackCount := int32(0)
	wg.Add(concurrentCalls)

	// Start multiple goroutines
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			ExecuteAsync(ctx, func() error {
				time.Sleep(10 * time.Millisecond)
				return errors.New("error")
			}, func() {
				atomic.AddInt32(&callbackCount, 1)
				done <- true
			}, func(err error) {
				atomic.AddInt32(&callbackCount, 1)
				done <- true
			}, WithAsyncMaxRetries(10), WithAsyncRetryDelay(20*time.Millisecond))

			// Wait for callback
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
		}()
	}

	// Cancel context while operations are running
	time.Sleep(5 * time.Millisecond)
	cancel()

	wg.Wait()
	// All operations should have triggered a callback
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&callbackCount), "All callbacks should be called")
}

func TestExecuteAsync_RaceCondition_CallbackExecution(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 100
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[int]bool)
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		id := i
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			ExecuteAsync(ctx, func() error {
				return nil
			}, func() {
				mu.Lock()
				results[id] = true
				mu.Unlock()
				done <- true
			}, func(err error) {
				t.Errorf("Unexpected failure for goroutine %d", id)
				done <- true
			}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(1*time.Millisecond))

			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}
		}()
	}

	wg.Wait()
	mu.Lock()
	assert.Equal(t, concurrentCalls, len(results), "All callbacks should be executed")
	mu.Unlock()
}

func TestExecuteAsync_RaceCondition_MultipleRetriesWithSharedCounter(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 20
	var wg sync.WaitGroup
	totalAttempts := int32(0)
	completedCount := int32(0)
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			done := make(chan bool, 1)
			attemptCount := int32(0)

			ExecuteAsync(ctx, func() error {
				atomic.AddInt32(&totalAttempts, 1)
				count := atomic.AddInt32(&attemptCount, 1)
				if count < 2 {
					return errors.New("temporary error")
				}
				return nil
			}, func() {
				atomic.AddInt32(&completedCount, 1)
				done <- true
			}, func(err error) {
				done <- true
			}, WithAsyncMaxRetries(3), WithAsyncRetryDelay(1*time.Millisecond))

			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&completedCount), "All operations should succeed")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&totalAttempts), int32(concurrentCalls*2), "Should have multiple attempts per operation")
}
