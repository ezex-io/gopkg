package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteSync_Success(t *testing.T) {
	callCount := 0

	err := ExecuteSync(t.Context(), func() error {
		callCount++

		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "Expected function to be called once")
}

func TestExecuteSync_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	expectedCalls := 2

	err := ExecuteSync(t.Context(), func() error {
		callCount++
		if callCount < expectedCalls {
			return errors.New("temporary error")
		}

		return nil
	}, WithSyncMaxRetries(3), WithSyncRetryDelay(10*time.Millisecond))

	require.NoError(t, err)
	assert.Equal(t, expectedCalls, callCount, "Expected function to be called %d times", expectedCalls)
}

func TestExecuteSync_AllRetriesFail(t *testing.T) {
	callCount := 0
	expectedError := errors.New("persistent error")
	maxRetries := 3

	err := ExecuteSync(t.Context(), func() error {
		callCount++

		return expectedError
	}, WithSyncMaxRetries(maxRetries), WithSyncRetryDelay(10*time.Millisecond))

	require.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, maxRetries, callCount)
}

func TestExecuteSync_ContextCancellationDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
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

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, callCount, 1, "Should be called at least once")
	assert.Less(t, callCount, 5, "Should not complete all retries")
}

func TestExecuteSync_ConcurrentCalls(t *testing.T) {
	concurrentCalls := 10
	var wg sync.WaitGroup
	wg.Add(concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			defer wg.Done()
			err := ExecuteSync(t.Context(), func() error {
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}
