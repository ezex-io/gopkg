package retry

import (
	"context"
	"sync"
	"time"
)

type AsyncOptions func(*asyncOptions)

type asyncOptions struct {
	maxRetries int
	retryDelay time.Duration
}

func _defaultAsyncOpts() *asyncOptions {
	return &asyncOptions{
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

func WithAsyncMaxRetries(maxRetries int) AsyncOptions {
	return func(o *asyncOptions) {
		o.maxRetries = maxRetries
	}
}

func WithAsyncRetryDelay(retryDelay time.Duration) AsyncOptions {
	return func(o *asyncOptions) {
		o.retryDelay = retryDelay
	}
}

// ExecuteAsync executes a function asynchronously with retry logic
// It respects context cancellation and timeout
// onSuccess and onFailure callbacks will be called exactly once
func ExecuteAsync(
	ctx context.Context,
	fn func() error,
	onSuccess func(),
	onFailure func(error),
	opts ...AsyncOptions,
) {
	conf := _defaultAsyncOpts()
	for _, opt := range opts {
		opt(conf)
	}

	var once sync.Once
	callSuccess := func() {
		once.Do(func() {
			if onSuccess != nil {
				onSuccess()
			}
		})
	}
	callFailure := func(err error) {
		once.Do(func() {
			if onFailure != nil {
				onFailure(err)
			}
		})
	}

	go func() {
		var lastErr error
		for attempt := 0; attempt < conf.maxRetries; attempt++ {
			select {
			case <-ctx.Done():
				callFailure(ctx.Err())
				return
			default:
			}

			err := fn()
			if err == nil {
				callSuccess()
				return
			}

			lastErr = err

			if attempt < conf.maxRetries-1 {
				select {
				case <-ctx.Done():
					callFailure(ctx.Err())
					return
				case <-time.After(conf.retryDelay):
				}
			}
		}

		// All retries exhausted
		callFailure(lastErr)
	}()
}
