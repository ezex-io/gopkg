package retry

import (
	"context"
	"time"
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
	task Task,
	onSuccess func(),
	onFailure func(error),
	opts ...AsyncOptions,
) {
	ExecuteAsyncT(ctx, func() (any, error) {
		return nil, task()
	}, func(any) {
		if onSuccess != nil {
			onSuccess()
		}
	}, func(err error) {
		if onFailure != nil {
			onFailure(err)
		}
	}, opts...)
}

// ExecuteAsyncT executes a function asynchronously with retry logic and returns a result
// It respects context cancellation and timeout
// onSuccess and onFailure callbacks will be called exactly once.
func ExecuteAsyncT[T any](
	ctx context.Context,
	task TaskT[T],
	onSuccess func(T),
	onFailure func(error),
	opts ...AsyncOptions,
) {
	conf := defaultAsyncOpts()
	for _, opt := range opts {
		opt(conf)
	}

	go func() {
		var result T
		var err error
		for attempt := 0; attempt < conf.maxRetries; attempt++ {
			result, err = task()
			if err == nil {
				if onSuccess != nil {
					onSuccess(result)
				}

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
