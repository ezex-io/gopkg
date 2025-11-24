package retry

import (
	"context"
	"time"
)

type (
	AsyncTask func()
)

type AsyncOptions func(*asyncOptions)

type asyncOptions struct {
	maxRetries int
	retryDelay time.Duration
}

func defaultAsyncOpts() *asyncOptions {
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
// onSuccess and onFailure callbacks will be called exactly once.
func ExecuteAsync(
	ctx context.Context,
	task SyncTask,
	onFailure func(error),
	opts ...AsyncOptions,
) {
	conf := defaultAsyncOpts()
	for _, opt := range opts {
		opt(conf)
	}

	go func() {
		var err error
		for attempt := 0; attempt < conf.maxRetries; attempt++ {
			err = task()
			if err == nil {
				return
			}

			if attempt < conf.maxRetries-1 {
				select {
				case <-ctx.Done():
					if onFailure != nil {
						onFailure(ctx.Err())
					}

					return

				case <-time.After(conf.retryDelay):
				}
			}
		}

		// All retries exhausted
		onFailure(err)
	}()
}
