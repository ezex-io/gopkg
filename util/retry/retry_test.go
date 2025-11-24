package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteSync_SuccessFirstTry(t *testing.T) {
	called := int32(0)
	err := ExecuteSync(context.Background(), func() error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), called)
}

func TestExecuteSync_RetryUntilSuccess(t *testing.T) {
	called := int32(0)
	err := ExecuteSync(context.Background(), func() error {
		if atomic.AddInt32(&called, 1) < 3 {
			return errors.New("fail")
		}
		return nil
	}, WithMaxAttempts(5))
	assert.NoError(t, err)
	assert.Equal(t, int32(3), called)
}

func TestExecuteSync_ExceedMaxAttempts(t *testing.T) {
	called := int32(0)
	err := ExecuteSync(context.Background(), func() error {
		atomic.AddInt32(&called, 1)
		return errors.New("fail")
	}, WithMaxAttempts(4))
	assert.Error(t, err)
	assert.Equal(t, int32(4), called)
}

func TestExecuteSync_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ExecuteSync(ctx, func() error {
		return errors.New("fail")
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestExecuteSyncWithPredicate_CustomPredicate(t *testing.T) {
	called := int32(0)
	shouldRetry := func(err error) bool {
		return err.Error() == "retryable"
	}
	err := ExecuteSyncWithPredicate(context.Background(), func() error {
		if atomic.AddInt32(&called, 1) < 2 {
			return errors.New("retryable")
		}
		return errors.New("fatal")
	}, shouldRetry, WithMaxAttempts(5))
	assert.Error(t, err)
	assert.Equal(t, int32(2), called)
	assert.Equal(t, "fatal", err.Error())
}

func TestExecuteAsync_SuccessAndFailureCallbacks(t *testing.T) {
	called := int32(0)
	var success, failure int32
	done := make(chan struct{})
	ExecuteAsync(context.Background(), func() error {
		if atomic.AddInt32(&called, 1) < 2 {
			return errors.New("fail")
		}
		return nil
	}, func() {
		atomic.AddInt32(&success, 1)
		close(done)
	}, func(err error) {
		atomic.AddInt32(&failure, 1)
		close(done)
	}, WithMaxAttempts(3))
	<-done
	assert.Equal(t, int32(2), called)
	assert.Equal(t, int32(1), success)
	assert.Equal(t, int32(0), failure)
}

func TestExecuteAsync_FailureCallback(t *testing.T) {
	called := int32(0)
	var success, failure int32
	ch := make(chan struct{})
	ExecuteAsync(context.Background(), func() error {
		atomic.AddInt32(&called, 1)
		return errors.New("fail")
	}, func() {
		atomic.AddInt32(&success, 1)
		close(ch)
	}, func(err error) {
		atomic.AddInt32(&failure, 1)
		close(ch)
	}, WithMaxAttempts(2))
	<-ch
	assert.Equal(t, int32(2), called)
	assert.Equal(t, int32(0), success)
	assert.Equal(t, int32(1), failure)
}

func TestBackoffStrategies(t *testing.T) {
	exp := ExponentialBackoff(100*time.Millisecond, 2, 1*time.Second)
	linear := LinearBackoff(50 * time.Millisecond)
	fixed := FixedBackoff(200 * time.Millisecond)
	none := NoBackoff()
	assert.True(t, exp(1) >= 0)
	assert.True(t, linear(2) == 100*time.Millisecond)
	assert.True(t, fixed(3) == 200*time.Millisecond)
	assert.True(t, none(5) == 0)
}

func TestWithTimeout(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := ExecuteSync(ctx, func() error {
		time.Sleep(200 * time.Millisecond)
		return errors.New("fail")
	}, WithTimeout(100*time.Millisecond), WithMaxAttempts(5))
	assert.Error(t, err)
	assert.True(t, time.Since(start) < 300*time.Millisecond)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestOnRetryCallback(t *testing.T) {
	calls := make([]struct {
		Attempt int
		Err     error
		Wait    time.Duration
	}, 0)
	err := ExecuteSync(context.Background(), func() error {
		return errors.New("fail")
	}, WithMaxAttempts(3), WithOnRetry(func(attempt int, lastErr error, nextWait time.Duration) {
		calls = append(calls, struct {
			Attempt int
			Err     error
			Wait    time.Duration
		}{attempt, lastErr, nextWait})
	}))
	assert.Error(t, err)
	assert.Len(t, calls, 2)
	for _, c := range calls {
		assert.NotNil(t, c.Err)
		assert.True(t, c.Wait >= 0)
	}
}

func TestRetryableError_InterfaceChecks(t *testing.T) {
	tempErr := &tempError{}
	timeoutErr := &timeoutError{}
	assert.True(t, RetryableError(tempErr))
	assert.True(t, RetryableError(timeoutErr))
	assert.False(t, RetryableError(errors.New("other")))
	assert.False(t, RetryableError(nil))
}

type tempError struct{}

func (t *tempError) Error() string   { return "temp" }
func (t *tempError) Temporary() bool { return true }

type timeoutError struct{}

func (t *timeoutError) Error() string { return "timeout" }
func (t *timeoutError) Timeout() bool { return true }
