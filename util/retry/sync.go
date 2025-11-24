package retry

import (
	"context"
	"time"
)

type Options func(*syncOptions)

type syncOptions struct {
	maxRetries int
	retryDelay time.Duration
}

func _defaultSyncOpts() *syncOptions {
	return &syncOptions{
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

func WithSyncMaxRetries(maxRetries int) Options {
	return func(o *syncOptions) {
		o.maxRetries = maxRetries
	}
}

func WithSyncRetryDelay(retryDelay time.Duration) Options {
	return func(o *syncOptions) {
		o.retryDelay = retryDelay
	}
}

// ExecuteSync executes a function synchronously with retry logic
// It respects context cancellation and timeout
// Returns nil if the function succeeds, or the last error if all retries are exhausted
func ExecuteSync(ctx context.Context, fn func() error, opts ...Options) error {
	conf := _defaultSyncOpts()
	for _, opt := range opts {
		opt(conf)
	}

	var lastErr error
	for attempt := 0; attempt < conf.maxRetries; attempt++ {
		// Check if context is cancelled before each attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if attempt < conf.maxRetries-1 {
			// Wait before retry, but respect context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(conf.retryDelay):
				// Continue to next retry
			}
		}
	}

	// All retries exhausted
	return lastErr
}
