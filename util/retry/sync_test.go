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

func TestExecuteSync_Success(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := ExecuteSync(ctx, func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "Expected function to be called once")
}

func TestExecuteSync_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	expectedCalls := 2

	err := ExecuteSync(ctx, func() error {
		callCount++
		if callCount < expectedCalls {
			return errors.New("temporary error")
		}
		return nil
	}, WithSyncMaxRetries(3), WithSyncRetryDelay(10*time.Millisecond))

	assert.NoError(t, err)
	assert.Equal(t, expectedCalls, callCount, "Expected function to be called %d times", expectedCalls)
}

func TestExecuteSync_AllRetriesFail(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	expectedError := errors.New("persistent error")
	maxRetries := 3

	err := ExecuteSync(ctx, func() error {
		callCount++
		return expectedError
	}, WithSyncMaxRetries(maxRetries), WithSyncRetryDelay(10*time.Millisecond))

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, maxRetries, callCount)
}

func TestExecuteSync_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	// Cancel context immediately
	cancel()

	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("should not succeed")
	}, WithSyncMaxRetries(3))

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, callCount)
}

func TestExecuteSync_ContextCancellationDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	err := ExecuteSync(ctx, func() error {
		callCount++
		if callCount == 1 {
			// Cancel after first attempt
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
		}
		return errors.New("temporary error")
	}, WithSyncMaxRetries(5), WithSyncRetryDelay(100*time.Millisecond))

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, callCount, 1, "Should be called at least once")
	assert.Less(t, callCount, 5, "Should not complete all retries")
}

func TestExecuteSync_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0

	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("temporary error")
	}, WithSyncMaxRetries(10), WithSyncRetryDelay(100*time.Millisecond))

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, callCount, 1, "Should be called at least once")
	assert.Less(t, callCount, 10, "Should not complete all retries due to timeout")
}

func TestExecuteSync_CustomRetryCount(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	maxRetries := 5

	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncMaxRetries(maxRetries), WithSyncRetryDelay(1*time.Millisecond))

	assert.Error(t, err)
	assert.Equal(t, maxRetries, callCount)
}

func TestExecuteSync_CustomRetryDelay(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	retryDelay := 50 * time.Millisecond

	start := time.Now()
	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncMaxRetries(3), WithSyncRetryDelay(retryDelay))
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Should take at least 2 * retryDelay (2 delays between 3 attempts)
	expectedMinDuration := 2 * retryDelay
	assert.GreaterOrEqual(t, elapsed, expectedMinDuration, "Expected at least %v elapsed time", expectedMinDuration)
}

func TestExecuteSync_DefaultOptions(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	// Use default syncOptions (3 retries, 2 second delay)
	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncRetryDelay(1*time.Millisecond)) // Override delay to make test faster

	assert.Error(t, err)
	// Default is 3 retries
	assert.Equal(t, 3, callCount, "Expected function to be called 3 times (default)")
}

func TestExecuteSync_NoRetryDelay(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	start := time.Now()
	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncMaxRetries(3), WithSyncRetryDelay(0))
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Should complete very quickly with no delay
	assert.Less(t, elapsed, 100*time.Millisecond, "Expected quick execution with no delay")
	assert.Equal(t, 3, callCount)
}

func TestExecuteSync_SingleRetry(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncMaxRetries(1), WithSyncRetryDelay(1*time.Millisecond))

	assert.Error(t, err)
	assert.Equal(t, 1, callCount)
}

func TestExecuteSync_MultipleOptions(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	maxRetries := 4
	retryDelay := 10 * time.Millisecond

	err := ExecuteSync(ctx, func() error {
		callCount++
		return errors.New("persistent error")
	}, WithSyncMaxRetries(maxRetries), WithSyncRetryDelay(retryDelay))

	assert.Error(t, err)
	assert.Equal(t, maxRetries, callCount)
}

// Race condition tests

func TestExecuteSync_RaceCondition_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	concurrentCalls := 100
	var wg sync.WaitGroup
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			err := ExecuteSync(ctx, func() error {
				return nil
			}, WithSyncMaxRetries(3), WithSyncRetryDelay(1*time.Millisecond))
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestExecuteSync_RaceCondition_SharedState(t *testing.T) {
	ctx := context.Background()
	var counter int32
	concurrentCalls := 50
	var wg sync.WaitGroup
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			err := ExecuteSync(ctx, func() error {
				// Use atomic operations to properly synchronize access to shared state
				atomic.AddInt32(&counter, 1)
				return nil
			}, WithSyncMaxRetries(1), WithSyncRetryDelay(0))
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	// Verify all operations completed successfully
	assert.Equal(t, int32(concurrentCalls), atomic.LoadInt32(&counter))
}

func TestExecuteSync_RaceCondition_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	concurrentCalls := 50
	var wg sync.WaitGroup
	wg.Add(concurrentCalls)

	// Start multiple goroutines
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			_ = ExecuteSync(ctx, func() error {
				time.Sleep(10 * time.Millisecond)
				return errors.New("error")
			}, WithSyncMaxRetries(10), WithSyncRetryDelay(20*time.Millisecond))
		}()
	}

	// Cancel context while operations are running
	time.Sleep(5 * time.Millisecond)
	cancel()

	wg.Wait()
}
