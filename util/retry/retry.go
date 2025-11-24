package retry

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

// BackoffStrategy defines how to calculate the wait duration between retries
type BackoffStrategy func(attempt int) time.Duration

// Config holds common retry configuration
type Config struct {
	// MaxAttempts is the maximum number of retry attempts (including initial attempt)
	// Default: 3
	MaxAttempts int

	// BackoffStrategy defines how to calculate wait time between retries
	// If nil, uses ExponentialBackoff with default parameters
	BackoffStrategy BackoffStrategy

	// OnRetry is called before each retry attempt
	OnRetry func(attempt int, lastErr error, nextWait time.Duration)

	// Timeout is the maximum total time allowed for all retry attempts
	// If zero, no timeout is applied
	Timeout time.Duration
}

// SyncOptions is an option for ExecuteSync
type SyncOptions func(*Config)

// AsyncOptions is an option for ExecuteAsync
type AsyncOptions func(*Config)

// NewRetryConfig returns a default RetryConfig
func NewRetryConfig() *Config {
	return &Config{
		MaxAttempts:     3,
		BackoffStrategy: ExponentialBackoff(100*time.Millisecond, 1.5, 30*time.Second),
	}
}

// WithMaxAttempts sets the maximum number of attempts
func WithMaxAttempts(attempts int) func(*Config) {
	return func(rc *Config) {
		if attempts > 0 {
			rc.MaxAttempts = attempts
		}
	}
}

// WithBackoffStrategy sets a custom backoff strategy
func WithBackoffStrategy(strategy BackoffStrategy) func(*Config) {
	return func(rc *Config) {
		if strategy != nil {
			rc.BackoffStrategy = strategy
		}
	}
}

// WithOnRetry sets the retry callback
func WithOnRetry(onRetry func(attempt int, lastErr error, nextWait time.Duration)) func(*Config) {
	return func(rc *Config) {
		rc.OnRetry = onRetry
	}
}

// WithTimeout sets the total timeout for retry operations
func WithTimeout(timeout time.Duration) func(*Config) {
	return func(rc *Config) {
		if timeout > 0 {
			rc.Timeout = timeout
		}
	}
}

var randSource = rand.NewSource(time.Now().UnixNano())
var randMutex sync.Mutex

// ExponentialBackoff returns an exponential backoff strategy with jitter
// initialDelay: initial wait duration
// multiplier: exponential multiplier (typically 1.5 or 2.0)
// maxDelay: maximum wait duration
func ExponentialBackoff(initialDelay time.Duration, multiplier float64, maxDelay time.Duration) BackoffStrategy {
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}

		// Calculate exponential delay: initialDelay * (multiplier ^ (attempt - 1))
		delay := time.Duration(float64(initialDelay) * math.Pow(multiplier, float64(attempt-1)))

		// Cap at maxDelay
		if delay > maxDelay {
			delay = maxDelay
		}

		// Add jitter: randomize between 0 and delay
		// This prevents thundering herd problem
		randMutex.Lock()
		jitter := time.Duration(randSource.Int63() % int64(delay))
		randMutex.Unlock()

		return delay/2 + jitter/2 // Ensure jitter doesn't exceed delay
	}
}

// LinearBackoff returns a linear backoff strategy
// increment: duration to add between each retry
func LinearBackoff(increment time.Duration) BackoffStrategy {
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}
		return time.Duration(attempt) * increment
	}
}

// FixedBackoff returns a fixed backoff strategy
func FixedBackoff(duration time.Duration) BackoffStrategy {
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}
		return duration
	}
}

// NoBackoff returns immediately without waiting
func NoBackoff() BackoffStrategy {
	return func(attempt int) time.Duration {
		return 0
	}
}

// ExecuteSync executes a function synchronously with retry logic
// It respects context cancellation and timeout
func ExecuteSync(
	ctx context.Context,
	fn func() error,
	opts ...SyncOptions,
) error {
	config := NewRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	return retryLoop(ctx, fn, config, nil)
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
	config := NewRetryConfig()
	for _, opt := range opts {
		opt(config)
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
		err := retryLoop(ctx, fn, config, nil)
		if err == nil {
			callSuccess()
		} else {
			callFailure(err)
		}
	}()
}

// ExecuteSyncWithPredicate executes with a predicate to determine if retry should occur
func ExecuteSyncWithPredicate(
	ctx context.Context,
	fn func() error,
	shouldRetry IsRetryable,
	opts ...SyncOptions,
) error {
	if shouldRetry == nil {
		shouldRetry = RetryableError
	}

	config := NewRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	return retryLoop(ctx, fn, config, shouldRetry)
}

// IsRetryable is a helper to determine if an error should trigger a retry
// You can implement custom logic based on error types or values
type IsRetryable func(error) bool

// RetryableError checks if an error is transient and should be retried
// Common transient errors: temporary network failures, timeouts, 5xx HTTP responses
func RetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Example: check for temporary interface
	if temp, ok := err.(interface{ Temporary() bool }); ok {
		return temp.Temporary()
	}

	// Example: check for timeout interface
	if timeout, ok := err.(interface{ Timeout() bool }); ok {
		return timeout.Timeout()
	}

	return false
}

// retryLoop is the core retry logic shared by all retry functions
func retryLoop(
	ctx context.Context,
	fn func() error,
	config *Config,
	shouldRetry IsRetryable,
) error {
	retryCtx := ctx
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		retryCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		select {
		case <-retryCtx.Done():
			return retryCtx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if shouldRetry != nil && !shouldRetry(err) {
			return err
		}

		if attempt == config.MaxAttempts-1 {
			return lastErr
		}

		waitDuration := config.BackoffStrategy(attempt)

		if config.OnRetry != nil {
			nextWait := config.BackoffStrategy(attempt + 1)
			config.OnRetry(attempt+1, lastErr, nextWait)
		}

		select {
		case <-time.After(waitDuration):
		case <-retryCtx.Done():
			return retryCtx.Err()
		}
	}

	return lastErr
}
